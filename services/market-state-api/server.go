package marketstateapi

import (
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, provider Provider) (*Server, error) {
	handler, err := NewHandler(provider)
	if err != nil {
		return nil, err
	}
	if addr == "" {
		return nil, fmt.Errorf("server address is required")
	}
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler.Routes(),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (s *Server) ListenAndServe() error {
	if s == nil || s.httpServer == nil {
		return fmt.Errorf("server is required")
	}
	return s.httpServer.ListenAndServe()
}
