package server

// import (
// 	p "startier/internal/node/common/protocol"

// 	"github.com/DarthPestilane/easytcp"
// )

// func (s *Server) handleInfo(c easytcp.Context) {
// 	req := p.InfoReq{}
// 	if err := c.Bind(&req); err != nil {
// 		c.SetResponse(p.MSG_ID_INFO_RES, &p.InfoRes{Code: 400})
// 		return
// 	}
// 	res := p.InfoRes{}
// 	c.SetResponse(p.MSG_ID_INFO_RES, &res)
// }
