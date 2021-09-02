package server

import (
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
	"github.com/victornm/gtonline/internal/profile"
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
		App struct {
			Addr string
		}

		Auth struct {
			Secret string
		}

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

func DefaultConfig() Config {
	c := Config{}
	// App Config
	c.App.Addr = ":8080"
	c.Auth.Secret = "JznqcOJCAEc1aq7Zulm83OtQt7md2gOK"

	// DB config
	c.DB.Addr = "mysql:3306"
	c.DB.User = "root"
	c.DB.Pass = "root"
	c.DB.Name = "gt-online"
	return c
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
	log.Printf("DB config: addr=%s, user=%s, name=%s", cfg.Addr, cfg.User, cfg.Name)

	stg, err := storage.NewWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("open db: %v", err)
	}

	if err := try(20, stg.Ping); err != nil {
		return fmt.Errorf("ping db: %v", err)
	}
	s.storage = stg

	return nil
}

func (s *Server) initRouter() {
	s.e = gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	s.e.Use(cors.New(corsConfig))
	s.e.GET("/ping", func(c *gin.Context) {
		c.JSON(200, map[string]string{"ping": "pong"})
	})

	a := &api.API{
		Auth:    auth.NewService(s.storage, []byte(s.cfg.Auth.Secret)),
		Profile: profile.NewService(s.storage),
	}
	a.Route(s.e)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.init()
	s.e.ServeHTTP(w, r)
}

func (s *Server) Start() {
	s.init()
	log.Println("Server start at", s.cfg.App.Addr)
	if err := s.e.Run(s.cfg.App.Addr); err != nil && err != http.ErrServerClosed {
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
