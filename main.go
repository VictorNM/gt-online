package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
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

	fmt.Println("Success")
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
