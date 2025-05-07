package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/proxy"
	"github.com/jesperkha/notifier"
)

type Server struct {
	mux    *chi.Mux
	config *config.Config
	db     *database.Database
}

func New(config *config.Config, db *database.Database) *Server {
	mux := chi.NewMux()

	p := proxy.New(config)
	if err := p.RegisterServices(db); err != nil {
		log.Fatal(err)
	}

	p.RegisterService(database.Service{
		Name: "foo",
		Port: "5500",
	})

	mux.Mount("/", p.Router())

	return &Server{
		mux:    mux,
		config: config,
		db:     db,
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
		finish()
	}()

	log.Println("listening on port " + s.config.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println(err)
	}
}
