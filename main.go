package main

import (
	"github.com/victornm/gtonline/internal/server"
)

func main() {
	s := server.New(server.DefaultConfig())
	s.Start()
}
