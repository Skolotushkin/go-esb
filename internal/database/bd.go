package database

import (
	"log"

	"go-esb/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func Connect(cfg *config.Config) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	log.Println("✅ Database connected")
	return db
}
