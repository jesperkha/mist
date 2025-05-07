package main

import (
	"log"
	"os"
	"syscall"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/server"
	"github.com/jesperkha/notifier"
)

func main() {
	config := config.Load()
	db := database.New(config)
	notif := notifier.New()

	server := server.New(config, db)
	go server.ListenAndServe(notif)

	notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("shutdown")
}
