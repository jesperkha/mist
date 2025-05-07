package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/proxy"
	"github.com/jesperkha/mist/service"
)

func proxyHandler(db *database.Database) http.Handler {
	p := proxy.New()
	if err := p.RegisterServices(db); err != nil {
		log.Fatal(err)
	}

	return p.Router()
}

func serviceHandler(monitor *service.Monitor) http.Handler {
	mux := chi.NewMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		services, err := monitor.Poll()
		if err != nil {
			http.Error(w, "failed to poll services", http.StatusInternalServerError)
			return
		}

		if err := JSON(w, services); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	return mux
}

func JSON(w http.ResponseWriter, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Write(b)
	w.Header().Add("Content-Type", "application/json")
	return nil
}
