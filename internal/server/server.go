package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/victornm/gtonline/internal/api"
	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/storage"
)

type (
	Server struct {
		cfg Config

		once    sync.Once
		storage *storage.Storage
		e       *gin.Engine
	}

	Config struct {
		DB struct {
			Addr string
			User string
			Pass string
			Name string
		}
	}
)

func New(cfg Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) init() {
	s.once.Do(func() {
		if err := s.initStorage(); err != nil {
			log.Fatalf("init storage: %v", err)
		}
		s.initRouter()
	})
}

func (s *Server) initStorage() error {
	cfg := s.cfg.DB
	log.Printf("DB config: addr=%s", cfg.Addr)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg.User, cfg.Pass, cfg.Addr, cfg.Name))
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	if err := try(20, db.Ping); err != nil {
		return fmt.Errorf("ping db: %v", err)
	}
	s.storage = storage.New(db)
	return nil
}

func (s *Server) initRouter() {
	s.e = gin.Default()
	s.e.Use(cors.Default())
	s.e.GET("/ping", func(c *gin.Context) {
		c.JSON(200, `{"ping": "pong"}`)
	})

	a := &api.API{
		Auth: auth.NewService(s.storage),
	}
	a.Route(s.e)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.init()
	s.e.ServeHTTP(w, r)
}

func (s *Server) Start() {
	s.init()
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
		log.Printf("failed: attempt %d, reason: %v. Sleep for %v", i+1, err, interval)
		time.Sleep(interval)
	}

	return fmt.Errorf("failed after trying with %d time(s): %w", times, err)
}
