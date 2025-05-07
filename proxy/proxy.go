package proxy

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jesperkha/mist/database"
)

type Proxy struct {
	r *chi.Mux
}

func New() *Proxy {
	return &Proxy{
		r: chi.NewRouter(),
	}
}

func (p *Proxy) RegisterServices(db *database.Database) error {
	all, err := db.GetAllServices()
	if err != nil {
		return err
	}

	for _, s := range all {
		p.register(s)
	}

	return nil
}

func (p *Proxy) Router() *chi.Mux {
	return p.r
}

func (p *Proxy) register(service database.Service) {
	url, err := url.Parse(serviceUrl(service.Port))
	if err != nil {
		log.Fatal(err)
	}

	endpoint := "/" + service.Name

	p.r.Handle(endpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectTo(r.URL, url)

		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			log.Println(err)
		}

		defer res.Body.Close()

		if _, err := io.Copy(w, res.Body); err != nil {
			log.Println(err)
		}

		maps.Copy(w.Header(), res.Header)
	}))
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
