package main

import (
	"log"

	"github.com/victornm/gtonline/internal/config"
	"github.com/victornm/gtonline/internal/server"
)

func main() {
	var (
		f   = "config/default.yaml"
		cfg server.Config
	)

	if err := config.Load(f, &cfg); err != nil {
		log.Fatalf("load config from file %s: %v", f, err)
	}

	server.New(cfg).Start()
}
