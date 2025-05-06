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

	db := database.New(config)

	p := proxy.New(config)
	if err := p.RegisterServices(db); err != nil {
		log.Fatal(err)
	}

	p.RegisterService(database.Service{
		Name: "foo",
		Port: "5500",
	})

	go p.ListenAndServe(notif)

	notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("shutdown")
}
