package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("addr", ":4000", "server address to listen on")
	environment := flag.String("environment", "debug", "logger level")
	flag.Parse()

	level, err := logrus.ParseLevel(*environment)
	if err != nil {
		fmt.Printf("Error parsing log level: %v\n", err)
		os.Exit(1)
	}

	logger := logrus.New()
	logger.SetLevel(level)

	s := NewServer(*addr, logger)

	go func() {
		if err := s.Start(); err != nil {
			s.logger.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	s.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Graceful shutdown completed")
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout, forcing exit")
	}
}
