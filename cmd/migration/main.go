package main

import (
	"database/sql"
	"log"
	"log/slog"
	"notes-go/internal/config"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	// load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		return
	}

	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatal(err)
	}
}
