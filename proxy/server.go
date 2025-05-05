package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
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

func (s *Server) Register(service database.Service) {
	url, err := url.Parse(serviceUrl(service.Port))
	if err != nil {
		log.Fatal(err)
	}

	endpoint := "/" + service.Name

	s.handle(endpoint, func(ctx *Context) int {
		redirectTo(ctx.r.URL, url)

		res, err := http.DefaultTransport.RoundTrip(ctx.r)
		if err != nil {
			log.Println(err)
			return http.StatusInternalServerError
		}

		defer res.Body.Close()

		if _, err := io.Copy(ctx.w, res.Body); err != nil {
			log.Println(err)
			return http.StatusInternalServerError
		}

		maps.Copy(ctx.w.Header(), res.Header)
		return res.StatusCode
	})
}

func serviceUrl(port string) string {
	if port[0] != ':' {
		port = ":" + port
	}
	return fmt.Sprintf("http://127.0.0.1%s", port)
}

func redirectTo(req *url.URL, to *url.URL) {
	req.Host = to.Host
	req.Scheme = to.Scheme
	req.Path = removeFirstPathSegment(req.Path)
}

func removeFirstPathSegment(path string) string {
	trimmed := strings.Trim(path, "/")
	split := strings.Split(trimmed, "/")

	if len(split) > 1 {
		return "/" + strings.Join(split[1:], "/")
	}

	return "/"
}

// Handle endpoint with handler wrapped with given middlewares.
func (s *Server) handle(pattern string, handler Handler, middlewares ...Middleware) {
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{w: w, r: r}
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
