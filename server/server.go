package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/service"
	"github.com/jesperkha/notifier"
)

type Server struct {
	mux     *chi.Mux
	config  *config.Config
	db      *database.Database
	cleanup func()
}

func New(config *config.Config, db *database.Database) *Server {
	mux := chi.NewMux()

	mux.Use(middleware.Logger)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	monitor := service.NewMonitor(db)

	mux.Mount("/", proxyHandler(db))
	mux.Mount("/service", serviceHandler(config, monitor))
	mux.Mount("/auth", authHandler(config))

	cleanup := func() {
		monitor.CloseConn()
	}

	return &Server{
		mux:     mux,
		config:  config,
		db:      db,
		cleanup: cleanup,
	}
}

func (s *Server) ListenAndServe(notif *notifier.Notifier) {
	done, finish := notif.Register()

	server := &http.Server{
		Addr:    s.config.Port,
		Handler: s.mux,
	}

	go func() {
		<-done
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}

		s.cleanup()
		finish()
	}()

	log.Println("listening on port " + s.config.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println(err)
	}
}
