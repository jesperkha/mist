package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/service"
)

func main() {
	config := config.Load()
	// notif := notifier.New()

	// // db := database.New(config)

	// s := proxy.New(config)
	// s.Use(proxy.Logger)

	// go s.ListenAndServe(notif)

	// notif.NotifyOnSignal(os.Interrupt, syscall.SIGTERM)
	// log.Println("shutdown")

	m := service.NewMonitor(config)
	u, err := m.Poll()
	if err != nil {
		log.Fatal(err)
	}

	for _, uu := range u {
		b, _ := json.MarshalIndent(uu, "", "  ")
		fmt.Println(string(b))
	}

	m.Close()

}
