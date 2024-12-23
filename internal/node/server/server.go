package server

import (
	"fmt"
	"startier/config"
	"startier/internal/node/client"
	p "startier/internal/node/common/protocol"
	"startier/internal/node/database"
	"startier/internal/node/tun"

	"github.com/DarthPestilane/easytcp"
)

type Server struct {
	config   *config.Config
	database *database.Database
	server   *easytcp.Server
	client   *client.Client
	tun      *tun.Tun
}
type Route struct {
	ID      interface{}
	Handler func(easytcp.Context)
}

func New(
	conf *config.Config,
	db *database.Database,
	cl *client.Client,
	t *tun.Tun,
) (*Server, error) {
	s := &Server{
		config:   conf,
		database: db,
		client:   cl,
		tun:      t,
	}
	s.server = easytcp.NewServer(&easytcp.ServerOption{
		Packer:      easytcp.NewDefaultPacker(),
		Codec:       &easytcp.ProtobufCodec{},
		AsyncRouter: true,
	})
	return s, nil
}

func (s *Server) Run(ch chan error) {
	s.LoadRoutes()
	addr := fmt.Sprintf("%s:%d", s.config.Server.Listen, s.config.Server.Port)
	if s.config.Server.TLS.Enable {
	} else {
		err := s.server.Run(addr)
		if err != nil {
			ch <- err
		}
	}
}

func (s *Server) LoadRoutes() {
	for _, r := range s.GetRoutes() {
		s.server.AddRoute(r.ID, r.Handler)
		fmt.Printf("%T %v\n", r.ID, r.ID)
	}
}

func (s *Server) GetRoutes() []Route {
	return []Route{
		{ID: p.MSG_ID_INFO_REQ, Handler: s.handleInfo},
		// {ID: p.MSG_ID_INFO_REQ, Handler: s.handle_info},
		// {ID: p.MSG_ID_INFO_REQ, Handler: s.handle_info},
		// {ID: p.MSG_ID_INFO_REQ, Handler: s.handle_info},
	}
}
