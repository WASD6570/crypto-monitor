package marketstateapi

import (
	"context"
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

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.httpServer == nil {
		return fmt.Errorf("server is required")
	}
	if ctx == nil {
		return fmt.Errorf("shutdown context is required")
	}
	return s.httpServer.Shutdown(ctx)
}
