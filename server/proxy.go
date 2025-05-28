package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
)

// Global map of proxy handler with the service name as key.
var proxyHandlers map[string]http.Handler

const ctxCookieName = "Mist-Service-Context"

func newProxyRouter(config *config.Config, db *database.Database) (*chi.Mux, error) {
	proxyHandlers = make(map[string]http.Handler)
	mux := chi.NewMux()

	all, err := db.GetAllServices()
	if err != nil {
		return nil, err
	}

	for _, s := range all {
		addProxyHandler(config, s)
	}

	mux.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		name := getFirstSegment(r)
		if h, ok := proxyHandlers[name]; ok {
			h.ServeHTTP(w, r)
			return
		}

		if cookie, err := r.Cookie(ctxCookieName); err == nil {
			if h, ok := proxyHandlers[cookie.Value]; ok {
				r.URL.Path = fmt.Sprintf("/%s%s", cookie.Value, r.URL.Path)
				h.ServeHTTP(w, r)
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return mux, nil
}

func getFirstSegment(r *http.Request) string {
	path := strings.Trim(r.URL.Path, "/")
	segments := strings.SplitN(path, "/", 2)
	if len(segments) > 0 && segments[0] != "" {
		return segments[0]
	}
	return ""
}

func addProxyHandler(config *config.Config, s database.Service) {
	h := makeProxyHandler(s)
	if s.RequireAuth {
		h = requireAuth(config)(h)
	}

	proxyHandlers[s.Name] = h
}

func removeProxyHandler(s database.Service) {
	delete(proxyHandlers, s.Name)
}

func makeProxyHandler(service database.Service) http.Handler {
	url, err := url.Parse(serviceUrl(service.Port))
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectTo(r.URL, url)

		if _, err := r.Cookie(ctxCookieName); err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:  ctxCookieName,
				Value: service.Name,
			})
		}

		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer res.Body.Close()

		for k, values := range res.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(res.StatusCode)

		if _, err := io.Copy(w, res.Body); err != nil {
			log.Println(err)
			return
		}
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
