package easynode

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetConfig() *Config {
	return _config
}

func GetDatabase() *gorm.DB {
	return _db
}

func GetServer() *Server {
	return _server
}

func GetLog() *zap.SugaredLogger {
	return _log
}

func GetJoinMessage() *JoinMessage {
	return _join_msg
}

func LoadStop[T any](getter func() *T) {
	for getter() == nil {
		time.Sleep(time.Millisecond * 100)
	}
}

func Load(v interface{}) {
	switch v.(type) {
	case *Config:
		LoadStop(GetConfig)
	case *gorm.DB:
		LoadStop(GetDatabase)
	case *Server:
		LoadStop(GetServer)
	case *zap.SugaredLogger:
		LoadStop(GetLog)
	case *JoinMessage:
		LoadStop(GetJoinMessage)
	default:
		panic("type is not in list")
	}
}

// func Load[T any](v *T) *T {
// 	for v == nil {
// 		time.Sleep(time.Microsecond * 100)
// 	}
// 	return v
// }
