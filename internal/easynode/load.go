package easynode

import (
	"crypto/tls"
	"crypto/x509"
	"os"
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
	default:
		panic("type is not in list")
	}
}

func loadTLSConfig(public, private, ca string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(public, private)
	if err != nil {
		return nil, err
	}
	var caCertPool *x509.CertPool
	if ca != "" {
		caCert, err := os.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
