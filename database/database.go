package database

import (
	"log"

	"github.com/jesperkha/mist/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func New(config *config.Config) *Database {
	conn, err := gorm.Open(sqlite.Open(config.BDPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.AutoMigrate(&Service{}); err != nil {
		log.Fatalf("migration: %v", err)
	}

	return &Database{
		conn: conn,
	}
}
