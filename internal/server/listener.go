package server

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	addr     string
	listener net.Listener
	mu       sync.Mutex
	conns    map[net.Conn]bool
	done     chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:  addr,
		conns: make(map[net.Conn]bool),
		done:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = listener
	fmt.Printf("Server listening on %s\n", s.addr)

	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	defer s.listener.Close()
	for {
		select {
		case <-s.done:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				fmt.Printf("Accept error: %v\n", err)
			}
			continue
		}

		s.mu.Lock()
		s.conns[conn] = true
		s.mu.Unlock()

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
	}()

	fmt.Printf("Client connected: %s\n", conn.RemoteAddr())
	handler := NewHandler(conn)
	handler.Run()
}

func (s *Server) Stop() {
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()
}
