package easynode

import (
	"context"
	"github.com/google/uuid"
	"io"
)

type Tunnel interface {
	ID() interface{}
	SetID(interface{})
	From() Session
	To() Session
}

var _ Tunnel = &tunnel{}

type tunnel struct {
	id   interface{}
	from Session
	to   Session
}

func newTunnel(from, to Session) *tunnel {
	return &tunnel{
		id:   uuid.NewString(),
		from: from,
		to:   to,
	}
}

func (t *tunnel) ID() interface{} {
	return t.id
}

func (t *tunnel) SetID(id interface{}) {
	t.id = id
}

func (t *tunnel) From() Session {
	return t.from
}

func (t *tunnel) To() Session {
	return t.to
}

func (t *tunnel) handle() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go t.pipe(ctx, t.from.Conn(), t.to.Conn())
	go t.pipe(ctx, t.to.Conn(), t.from.Conn())
	<-ctx.Done()
	if err := ctx.Err(); err != nil {
		_log.Errorf("tunnel %s is closed : %s", t.id, err)
	}
}

func (t *tunnel) pipe(ctx context.Context, src, dst io.ReadWriter) {
	_, err := io.Copy(dst, src)
	if err != nil {
		ctx.Done()
	}
}


// ----------

