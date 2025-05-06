package server

import (
	"context"
	"log"
	"net/http"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/notifier"
)

type Server struct {
	mux    *http.ServeMux
	config *config.Config
	h      http.Handler
}

// Handler takes in a Context with the request and response writer, and
// returns the status code after handling the request.
type Handler func(ctx *Context) int

func New(config *config.Config) *Server {
	mux := http.NewServeMux()
	s := &Server{
		mux:    mux,
		config: config,
		h:      mux,
	}

	return s
}

// Handle endpoint with handler wrapped with given middlewares.
func (s *Server) Handle(pattern string, handler Handler, middlewares ...Middleware) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{W: w, R: r}
		code := handler(ctx)

		if code != http.StatusOK {
			w.WriteHeader(code)
		}
	})

	h := http.Handler(hf) // Cast handler

	// Wrap middlewares
	for _, m := range middlewares {
		h = m(h)
	}

	s.mux.Handle(pattern, h)
}

func (s *Server) ListenAndServe(notif *notifier.Notifier) {
	done, finish := notif.Register()

	server := &http.Server{
		Addr:    s.config.Port,
		Handler: s.h,
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}
