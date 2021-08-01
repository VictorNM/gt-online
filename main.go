package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	s := &Server{}
	s.init()
	s.start()
}

type Server struct {
	db *sql.DB
	e  *gin.Engine
}

func (s *Server) init() {
	var (
		user, pass = "root", "root"
		addr       = "mysql:3306"
		name       = "gt-online"
	)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, addr, name))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := try(20, db.Ping); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	s.db = db

	s.e = gin.Default()
	s.e.Use(cors.Default())
	s.e.GET("/ping", func(c *gin.Context) {
		c.JSON(200, `{"ping": "pong"}`)
	})
}

func (s *Server) start() {
	log.Println("Server start at port 8080")
	if err := s.e.Run(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Run server failed: %v", err)
	}
}

func try(times int, f func() error) error {
	var err error
	for i := 0; i < times; i++ {
		err = f()
		if err == nil {
			return nil
		}
		interval := time.Duration(i+1) * time.Second
		log.Printf("failed: attempt %d, sleep for %v", i+1, interval)
		time.Sleep(interval)
	}

	return fmt.Errorf("failed after trying with %d time(s): %w", times, err)
}
