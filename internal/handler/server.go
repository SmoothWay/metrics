package handler

import (
	"context"
	"net/http"
)

type server struct {
	server *http.Server
}

func NewServer(host string, h *Handler, key, trustedSubnet string, privateKey []byte) *server {
	s := &http.Server{
		Addr:    host,
		Handler: Router(h, key, trustedSubnet, privateKey),
	}
	return &server{
		server: s,
	}
}

func (s *server) Run() error {
	return s.server.ListenAndServe()

}

func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
