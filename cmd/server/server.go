package main

import (
	"log/slog"
	"net"
	"redisclone/internal/storage"
	"sync"
)

type Server struct {
	addr    string
	storage *storage.Storage
	mu      sync.RWMutex
	clients map[net.Conn]bool
	closeCh chan struct{}
	doneCh  chan struct{}
}

func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		storage: storage.NewStorage(),
		clients: make(map[net.Conn]bool),
		closeCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer l.Close()

	slog.Info("Server started", "addr", s.addr)

	go func() {
		<-s.closeCh
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.closeCh:
				close(s.doneCh)
				return nil
			default:
				slog.Error("accept error", "err", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}

}

func (s *Server) Stop() {
	close(s.closeCh)
	<-s.doneCh

	for conn := range s.clients {
		conn.Close()
	}
}
