package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/proxy"
)

func proxyHandler(config *config.Config, db *database.Database) http.Handler {
	p := proxy.New(config)
	if err := p.RegisterServices(db); err != nil {
		log.Fatal(err)
	}

	p.RegisterService(database.Service{
		Name: "foo",
		Port: "5500",
	})

	return p.Router()
}

func serviceHandler(config *config.Config, db *database.Database) http.Handler {
	mux := chi.NewMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello\n"))
	})

	return mux
}
