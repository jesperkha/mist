package database

import (
	"log"

	"github.com/jesperkha/mist/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func New(config *config.Config) *Database {
	db, err := gorm.Open(sqlite.Open(config.BDPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("migration: %v", err)
	}

	return &Database{
		db: db,
	}
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate()
}
