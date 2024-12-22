package web

import (
	"startier/config"
	"startier/internal/node/database"
)

type Server struct {
	config   *config.Config
	database *database.Database
}

func New(conf *config.Config, db *database.Database) (*Server, error) {
	return &Server{
		config:   conf,
		database: db,
	}, nil
}

func (s *Server) Run(ch chan error) {
}
