package server

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
)

func newProxyRouter(config *config.Config, db *database.Database) (*chi.Mux, error) {
	mux := chi.NewMux()

	all, err := db.GetAllServices()
	if err != nil {
		return nil, err
	}

	for _, s := range all {
		endpoint := "/" + s.Name
		handler := makeProxyHandler(s)

		if s.RequireAuth {
			handler = requireAuth(config)(handler)
		}

		mux.Handle(endpoint, handler)
	}

	return mux, nil
}

func makeProxyHandler(service database.Service) http.Handler {
	url, err := url.Parse(serviceUrl(service.Port))
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectTo(r.URL, url)

		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer res.Body.Close()

		if _, err := io.Copy(w, res.Body); err != nil {
			log.Println(err)
			return
		}

		maps.Copy(w.Header(), res.Header)
	})
}

func serviceUrl(port string) string {
	if !strings.HasPrefix(port, ":") {
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
