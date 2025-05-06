package proxy

import (
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
	"github.com/jesperkha/mist/server"
	"github.com/jesperkha/notifier"
)

type Proxy struct {
	s *server.Server
}

func New(config *config.Config) *Proxy {
	s := server.New(config)
	s.Use(server.Logger)

	return &Proxy{
		s: s,
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

func (p *Proxy) RegisterService(s database.Service) {
	p.register(s)
}

func (p *Proxy) ListenAndServe(notif *notifier.Notifier) {
	p.s.ListenAndServe(notif)
}

func (p *Proxy) register(service database.Service) {
	url, err := url.Parse(serviceUrl(service.Port))
	if err != nil {
		log.Fatal(err)
	}

	endpoint := "/" + service.Name

	p.s.Handle(endpoint, func(ctx *server.Context) int {
		redirectTo(ctx.R.URL, url)

		res, err := http.DefaultTransport.RoundTrip(ctx.R)
		if err != nil {
			log.Println(err)
			return http.StatusInternalServerError
		}

		defer res.Body.Close()

		if _, err := io.Copy(ctx.W, res.Body); err != nil {
			log.Println(err)
			return http.StatusInternalServerError
		}

		maps.Copy(ctx.W.Header(), res.Header)
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
