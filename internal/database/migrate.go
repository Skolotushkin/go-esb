package database

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB) {
	migrationDir := "migrations"
	
	// Читаем файлы миграций из директории
	migrationFiles, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Printf("⚠️ Warning: Could not read migrations directory: %v", err)
		log.Println("⚠️ Skipping migrations - make sure database schema is up to date")
		return
	}

	// Выполняем каждую миграцию
	for _, file := range migrationFiles {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		migrationPath := filepath.Join(migrationDir, file.Name())
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			log.Printf("⚠️ Could not read migration file %s: %v", file.Name(), err)
			continue
		}

		log.Printf("📝 Running migration: %s", file.Name())
		
		// Выполняем миграцию
		if _, err := db.Exec(string(content)); err != nil {
			// Игнорируем ошибки о существующих объектах
			errStr := err.Error()
			if strings.Contains(errStr, "already exists") || 
			   strings.Contains(errStr, "duplicate key") ||
			   strings.Contains(errStr, "already in") {
				log.Printf("ℹ️ Migration %s: objects already exist (skipped)", file.Name())
			} else {
				log.Printf("⚠️ Warning: Migration %s returned error: %v", file.Name(), err)
			}
		}
	}

	log.Println("✅ Migrations check completed")
}
