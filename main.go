package main

import (
	"flag"
	"log"

	"github.com/MahdiNajafzadeh/easynode/internal/easynode"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
}

func main() {
	log.Fatal(easynode.Run(configPath))
}
