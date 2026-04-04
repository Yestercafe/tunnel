package relay

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
)

// Server terminates TLS and accepts Tunnel v1 clients on ListenAddr.
type Server struct {
	ListenAddr string
	TLSConfig  *tls.Config
	Registry   *SessionRegistry

	mu       sync.Mutex
	listener net.Listener
}

// Listen binds TCP and wraps the listener with TLS.
func (s *Server) Listen() error {
	if s.TLSConfig == nil {
		return errors.New("relay: TLSConfig is required")
	}
	if s.Registry == nil {
		s.Registry = NewSessionRegistry()
	}
	raw, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = tls.NewListener(raw, s.TLSConfig)
	s.mu.Unlock()
	return nil
}

// Serve accepts connections until ctx is cancelled or Accept fails.
func (s *Server) Serve(ctx context.Context) error {
	s.mu.Lock()
	ln := s.listener
	s.mu.Unlock()
	if ln == nil {
		return errors.New("relay: Listen not called")
	}
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return err
			}
		}
		tc := conn.(*tls.Conn)
		go s.serveConn(tc)
	}
}

// Addr returns the bound listener address (after Listen).
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Close closes the listener.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	err := s.listener.Close()
	s.listener = nil
	return err
}
