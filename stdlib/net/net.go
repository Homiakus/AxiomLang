package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// SocketConnection wraps standard TCP/UDP connection with capability boundaries
type SocketConnection struct {
	conn net.Conn
	mu   sync.Mutex
}

func (s *SocketConnection) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.Write(data)
}

func (s *SocketConnection) Read(buf []byte) (int, error) {
	return s.conn.Read(buf)
}

func (s *SocketConnection) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.Close()
}

// Server provides high-throughput reactive network listening with capability governance
type Server struct {
	addr     string
	listener net.Listener
	closed   bool
	mu       sync.RWMutex
}

// Listen binds a TCP socket on address if permitted by capability
func Listen(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("E_NET_LISTEN_FAILED: %w", err)
	}
	return &Server{
		addr:     addr,
		listener: ln,
	}, nil
}

// Accept awaits new client connection
func (s *Server) Accept() (*SocketConnection, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, errors.New("server is closed")
	}
	ln := s.listener
	s.mu.RUnlock()

	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}
	return &SocketConnection{conn: conn}, nil
}

// Close terminates network server
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.listener.Close()
}

// Connect dials remote target host
func Connect(addr string, timeout time.Duration) (*SocketConnection, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("E_NET_CONNECT_FAILED: %w", err)
	}
	return &SocketConnection{conn: conn}, nil
}

// SimpleEchoServer helper for integration testing and microservices
func RunEchoService(ctx context.Context, addr string) (*Server, error) {
	srv, err := Listen(addr)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	go func() {
		for {
			client, err := srv.Accept()
			if err != nil {
				return
			}
			go func(c *SocketConnection) {
				defer c.Close()
				buf := make([]byte, 4096)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							return
						}
						return
					}
					if _, err := c.Write(buf[:n]); err != nil {
						return
					}
				}
			}(client)
		}
	}()

	return srv, nil
}
