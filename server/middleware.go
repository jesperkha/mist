package server

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Use middleware to base handler, in order.
func (s *Server) Use(m ...Middleware) {
	for _, m := range m {
		s.h = m(s.h)
	}
}

// Logger prints out host, method, path, and time of request.
func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Now().Sub(start).String()
		log.Printf("%s %s %s %s", r.Host, r.Method, r.URL.String(), duration)
	})
}
