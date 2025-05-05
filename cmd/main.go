package main

import (
	"log"
	"os"
	"syscall"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/proxy"
	"github.com/jesperkha/notifier"
)

func main() {
	config := config.Load()
	notif := notifier.New()

	// db := database.New(config)

	s := proxy.New(config)

	s.Use(proxy.Logger)
	s.Register(database.Service{
		Name: "foo",
		Port: "5500",
	})

	go s.ListenAndServe(notif)

	notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("shutdown")
}
