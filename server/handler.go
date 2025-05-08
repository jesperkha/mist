package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
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

func serviceHandler(config *config.Config, monitor *service.Monitor) http.Handler {
	mux := chi.NewMux()

	mux.Use(requireAuth(config))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		services, err := monitor.Poll()
		if err != nil {
			http.Error(w, "failed to poll services", http.StatusInternalServerError)
			return
		}

		templ := template.Must(template.ParseFiles("web/services.html"))
		templ.Execute(w, services)
	})

	return mux
}

func authHandler(config *config.Config) http.Handler {
	mux := chi.NewMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		origin := r.URL.Query().Get("origin")
		if origin == "" {
			origin = "/"
		}

		templ := template.Must(template.ParseFiles("web/auth.html"))
		templ.Execute(w, fmt.Sprintf("/auth?origin=%s", origin))
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		origin := r.URL.Query().Get("origin")
		password := r.FormValue("password")

		sum := sha256.Sum256([]byte(password))
		passHash := fmt.Sprintf("%x", sum)

		// Military grade auth
		if passHash != config.Secret {
			w.Write([]byte("incorrect password"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "AuthToken",
			Value: config.Secret,
		})

		w.Header().Set("HX-Redirect", origin)
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func requireAuth(config *config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("AuthToken")

			if err != nil || cookie.Value != config.Secret {
				log.Println("directing to auth")
				redirect := fmt.Sprintf("/auth?origin=%s", r.URL.Path)
				http.Redirect(w, r, redirect, http.StatusSeeOther)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
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
