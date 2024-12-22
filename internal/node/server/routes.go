package server

import (
	p "startier/internal/node/common/protocol"

	"github.com/DarthPestilane/easytcp"
)

func (s *Server) _InfoHandler(c easytcp.Context) {
	var req p.InfoReq
	if err := c.Bind(&req); err != nil {
		println("sss")
		c.SetResponse(
			p.MSG_ID_INFO_RES_ID,
			p.InfoRes{Code: 400},
		)
	}
	c.SetResponse(p.MSG_ID_INFO_RES_ID, p.InfoRes{Code: 200, Nodes: []*p.Node{}})
}
func (s *Server) _PacketHandler(c easytcp.Context) {}
func (s *Server) _TunnelHandler(c easytcp.Context) {}
