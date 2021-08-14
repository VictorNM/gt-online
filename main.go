package main

import (
	"github.com/victornm/gtonline/internal/server"
)

func main() {
	cfg := server.Config{}
	cfg.DB.Addr = "mysql:3306"
	cfg.DB.User = "root"
	cfg.DB.Pass = "root"
	cfg.DB.Name = "gt-online"

	s := server.New(cfg)
	s.Start()
}
