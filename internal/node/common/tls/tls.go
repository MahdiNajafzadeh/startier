package tls

import (
	"crypto/tls"
	"errors"
	"os"
	"startier/config"
)

func FileExists(file string) (bool, error) {
	info, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, errors.New("file is a directory")
	}
	return true, nil
}

func GetTLSConfig(c *config.Config) (*tls.Config, error) {
	if !c.Server.TLS.Enable {
		return nil, errors.New("TLS is not enabled")	
	}
	if c.Server.TLS.Files.Private == "" || c.Server.TLS.Files.Public == "" {
		return nil, errors.New("private or public key is empty")
	}
	private, err := FileExists(c.Server.TLS.Files.Private)
	if err != nil {
		return nil, err
	}
	if !private {
		return nil, errors.New("private key does not exist")
	}
	public, err := FileExists(c.Server.TLS.Files.Public)
	if err != nil {
		return nil, err
	}
	if !public {
		return nil, errors.New("public key does not exist")
	}
	cert, err := tls.LoadX509KeyPair(c.Server.TLS.Files.Public, c.Server.TLS.Files.Private)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
