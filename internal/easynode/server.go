package easynode

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -destination internal/mock/server_mock.go -package mock net Listener,Error,Conn

// Server is a server for TCP connections.
type Server struct {
	Listener net.Listener

	// Packer is the message packer, will be passed to session.
	Packer Packer

	// Codec is the message codec, will be passed to session.
	Codec Codec

	// OnSessionCreate is an event hook, will be invoked when session's created.
	OnSessionCreate func(sess Session)

	// OnSessionClose is an event hook, will be invoked when session's closed.
	OnSessionClose func(sess Session)

	socketReadBufferSize  int
	socketWriteBufferSize int
	socketSendDelay       bool
	readTimeout           time.Duration
	writeTimeout          time.Duration
	respQueueSize         int
	router                *Router
	printRoutes           bool
	acceptingC            chan struct{}
	stoppedC              chan struct{}
	asyncRouter           bool
	tls                   bool
	tlsConfig             *tls.Config
	sessionStore          Store[any, Session]
}

// ServerOption is the option for Server.
type ServerOption struct {
	NodeID                interface{}   // Unique ID for Node (deafult set UUID as NodeID)
	SocketReadBufferSize  int           // sets the socket read buffer size.
	SocketWriteBufferSize int           // sets the socket write buffer size.
	SocketSendDelay       bool          // sets the socket delay or not.
	ReadTimeout           time.Duration // sets the timeout for connection read.
	WriteTimeout          time.Duration // sets the timeout for connection write.
	Packer                Packer        // packs and unpacks packet payload, default packer is the DefaultPacker.
	Codec                 Codec         // encodes and decodes the message data, can be nil.
	RespQueueSize         int           // sets the response channel size of session, DefaultRespQueueSize will be used if < 0.
	DoNotPrintRoutes      bool          // whether to print registered route handlers to the console.

	// AsyncRouter represents whether to execute a route HandlerFunc of each session in a goroutine.
	// true means execute in a goroutine.
	AsyncRouter bool
}

// ErrServerStopped is returned when server stopped.
var ErrServerStopped = fmt.Errorf("server stopped")

const DefaultRespQueueSize = 1024

// NewServer creates a Server according to opt.
func NewServer(opt *ServerOption) *Server {
	if opt.Packer == nil {
		opt.Packer = NewDefaultPacker()
	}
	if opt.RespQueueSize < 0 {
		opt.RespQueueSize = DefaultRespQueueSize
	}
	if opt.NodeID == nil {
		opt.NodeID = uuid.NewString()
	}
	return &Server{
		socketReadBufferSize:  opt.SocketReadBufferSize,
		socketWriteBufferSize: opt.SocketWriteBufferSize,
		socketSendDelay:       opt.SocketSendDelay,
		respQueueSize:         opt.RespQueueSize,
		readTimeout:           opt.ReadTimeout,
		writeTimeout:          opt.WriteTimeout,
		Packer:                opt.Packer,
		Codec:                 opt.Codec,
		printRoutes:           !opt.DoNotPrintRoutes,
		router:                newRouter(),
		acceptingC:            make(chan struct{}),
		stoppedC:              make(chan struct{}),
		asyncRouter:           opt.AsyncRouter,
		tls:                   false,
		tlsConfig:             nil,
		sessionStore:          newStore[any, Session](),
	}
}

// Serve starts to serve the lis.
func (s *Server) Serve(lis net.Listener) error {
	s.Listener = lis
	if s.printRoutes {
		s.router.printHandlers(fmt.Sprintf("tcp://%s", s.Listener.Addr()))
	}
	return s.acceptLoop()
}

// Run starts to listen TCP and keeps accepting TCP connection in a loop.
// The loop breaks when error occurred, and the error will be returned.
func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}

// RunTLS starts serve TCP with TLS.
func (s *Server) RunTLS(addr string, config *tls.Config) error {
	lis, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	s.tls = true
	s.tlsConfig = config
	return s.Serve(lis)
}

// acceptLoop accepts TCP connections in a loop, and handle connections in goroutines.
// Returns error when error occurred.
func (s *Server) acceptLoop() error {
	close(s.acceptingC)
	for {
		if s.isStopped() {
			_log.Debugf("server accept loop stopped")
			return ErrServerStopped
		}

		conn, err := s.Listener.Accept()
		if err != nil {
			if s.isStopped() {
				_log.Debugf("server accept loop stopped")
				return ErrServerStopped
			}
			return fmt.Errorf("accept err: %s", err)
		}
		if s.socketReadBufferSize > 0 {
			if c, ok := conn.(*net.TCPConn); ok {
				if err := c.SetReadBuffer(s.socketReadBufferSize); err != nil {
					return fmt.Errorf("conn set read buffer err: %s", err)
				}
			}
		}
		if s.socketWriteBufferSize > 0 {
			if c, ok := conn.(*net.TCPConn); ok {
				if err := c.SetWriteBuffer(s.socketWriteBufferSize); err != nil {
					return fmt.Errorf("conn set write buffer err: %s", err)
				}
			}
		}
		if s.socketSendDelay {
			if c, ok := conn.(*net.TCPConn); ok {
				if err := c.SetNoDelay(false); err != nil {
					return fmt.Errorf("conn set no delay err: %s", err)
				}
			}
		}
		go s.handleConn(conn)
	}
}

// handleConn creates a new session with `conn`,
// handles the message through the session in different goroutines,
// and waits until the session's closed, then close the `conn`.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close() // nolint

	sess := newSession(conn, &sessionOption{
		Packer:        s.Packer,
		Codec:         s.Codec,
		respQueueSize: s.respQueueSize,
		asyncRouter:   s.asyncRouter,
	})

	s.Sessions().Set(sess.ID(), sess)

	if s.OnSessionCreate != nil {
		s.OnSessionCreate(sess)
	}
	close(sess.afterCreateHookC)

	go sess.readInbound(s.router, s.readTimeout) // start reading message packet from connection.
	go sess.writeOutbound(s.writeTimeout)        // start writing message packet to connection.

	select {
	case <-sess.closedC: // wait for session finished.
	case <-s.stoppedC: // or the server is stopped.
	}

	if s.OnSessionClose != nil {
		s.OnSessionClose(sess)
	}
	close(sess.afterCloseHookC)

	s.Sessions().Del(sess.ID())
}

// Stop stops server. Closing Listener and all connections.
func (s *Server) Stop() error {
	close(s.stoppedC)
	return s.Listener.Close()
}

// AddRoute registers message handler and middlewares to the router.
func (s *Server) AddRoute(msgID interface{}, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	s.router.register(msgID, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (s *Server) Use(middlewares ...MiddlewareFunc) {
	s.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (s *Server) NotFoundHandler(handler HandlerFunc) {
	s.router.setNotFoundHandler(handler)
}

func (s *Server) isStopped() bool {
	select {
	case <-s.stoppedC:
		return true
	default:
		return false
	}
}

func (s *Server) Connect(addr string) (net.Conn, error) {
	for s == nil {
		time.Sleep(time.Millisecond * 100)
	}
	var conn net.Conn
	var err error
	if s.tls {
		conn, err = tls.Dial("tcp", addr, s.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err == nil || conn != nil {
		go s.handleConn(conn)
	}
	return conn, err
}

func (s *Server) Request(addr string, id interface{}, v interface{}) error {
	conn, err := s.Connect(addr)
	if err != nil {
		return err
	}
	data, err := s.Codec.Encode(v)
	if err != nil {
		conn.Close()
		return err
	}
	packet, err := s.Packer.Pack(NewMessage(id, data))
	if err != nil {
		conn.Close()
		return err
	}
	_, err = conn.Write(packet)
	if err != nil {
		conn.Close()
		return err
	}
	go s.handleConn(conn)
	return nil
}

func (s *Server) Sessions() Store[any, Session] {
	return s.sessionStore
}

func (s *Server) BroadCast(id interface{}, v interface{}) {
	for _, sess := range s.sessionStore.All() {
		c := sess.AllocateContext()
		c.SetResponse(id, v)
		c.Send()
	}
}
