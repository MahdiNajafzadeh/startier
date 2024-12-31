package easynode

import "github.com/google/uuid"

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

func (t *tunnel) handleSess() {
	// for {

	// }
}
