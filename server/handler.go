package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/service"
)

func proxyHandler(config *config.Config, db *database.Database) http.Handler {
	mux, err := newProxyRouter(config, db)
	if err != nil {
		log.Fatal(err)
	}

	return mux
}

func dashboardHandler(config *config.Config, monitor *service.Monitor) http.Handler {
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

func serviceHandler(config *config.Config, db *database.Database, monitor *service.Monitor) http.Handler {
	mux := chi.NewMux()
	mux.Use(requireAuth(config))

	// Get all services
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		all, err := monitor.Poll()
		if err != nil {
			http.Error(w, "failed to poll", http.StatusInternalServerError)
			return
		}

		if err := JSON(w, all); err != nil {
			http.Error(w, "failed to write json", http.StatusInternalServerError)
		}
	})

	// Get service by ID
	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		s, err := db.GetServiceByID(uint(id))
		if err != nil {
			http.Error(w, "no service by that id", http.StatusNotFound)
			return
		}

		if err := JSON(w, s); err != nil {
			http.Error(w, "failed to write json", http.StatusInternalServerError)
		}
	})

	// Register new service from request body. Must be valid database.Service
	// object. Responds with new services id on success.
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Println(string(body))

		var s database.Service
		if err := json.Unmarshal(body, &s); err != nil || !ensureService(s) {
			http.Error(w, "bad data", http.StatusBadRequest)
			return
		}

		id, err := db.NewService(s)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Conent-Type", "text")
		w.Write(fmt.Appendf(nil, "%d", id))
	})

	// Delete service
	mux.HandleFunc("DELETE /{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		if err := db.RemoveService(uint(id)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	return mux
}

func ensureService(s database.Service) bool {
	if _, err := strconv.Atoi(s.Port); err != nil {
		return false
	}
	return s.Name != "" && s.Port != ""
}

func requireAuth(config *config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For browser clients
			cookie, err := r.Cookie("AuthToken")
			if err == nil && cookie.Value == config.Secret {
				h.ServeHTTP(w, r)
				return
			}

			// For API clients
			hdr := r.Header.Get("Authorization")
			if hdr == config.Secret {
				h.ServeHTTP(w, r)
				return
			}

			if isBrowserRequest(r) {
				redirect := fmt.Sprintf("/auth?origin=%s", r.URL.Path)
				http.Redirect(w, r, redirect, http.StatusSeeOther)
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
		})
	}
}

func isBrowserRequest(r *http.Request) bool {
	userAgent := r.Header.Get("User-Agent")
	accept := r.Header.Get("Accept")

	isUA := strings.Contains(userAgent, "Mozilla") ||
		strings.Contains(userAgent, "Chrome") ||
		strings.Contains(userAgent, "Safari") ||
		strings.Contains(userAgent, "Firefox")

	isAcceptsHTML := strings.Contains(accept, "text/html")
	return isUA || isAcceptsHTML
}

func JSON(w http.ResponseWriter, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(b)
	return err
}
