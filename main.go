package main

import (
	"flag"
	"log"
	"startier/internal/startier"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()
}

func main() {
	log.Fatal(startier.Run(configPath))
}