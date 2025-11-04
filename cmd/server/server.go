package main

import (
	"net"
	"redisclone/internal/storage"
	"sync"

	"github.com/sirupsen/logrus"
)

type Server struct {
	addr    string
	storage *storage.Storage
	mu      sync.RWMutex
	clients map[net.Conn]bool
	logger  *logrus.Logger
	closeCh chan struct{}
	doneCh  chan struct{}
}

func NewServer(addr string, logger *logrus.Logger) *Server {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
		})
	}

	return &Server{
		addr:    addr,
		storage: storage.NewStorage(),
		clients: make(map[net.Conn]bool),
		logger:  logger,
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

	s.logger.WithField("addr", s.addr).Info("Server started")

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
				s.logger.WithError(err).Error("accept error")
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
