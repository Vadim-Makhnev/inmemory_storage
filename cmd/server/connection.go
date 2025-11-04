package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"redisclone/internal/protocol"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *Server) handleConnection(conn net.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	remoteAddr := conn.RemoteAddr().String()
	s.logger.WithField("remote_addr", remoteAddr).Info("Client connected")

	defer func() {
		conn.Close()

		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()

		s.logger.WithField("remote_addr", remoteAddr).Info("Client disconnected")
	}()

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.WithError(err).Debug("Client disconnected")
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				s.logger.Warn("Read timeout, closing connection")
			} else {
				s.logger.WithError(err).Error("Read error")
			}
			return
		}

		line = strings.TrimSpace(line)

		cmd, err := protocol.ParseCommand(line)
		if err != nil {
			conn.Write([]byte(protocol.Error("bad command")))
			continue
		}

		if cmd == nil {
			continue
		}

		response := s.executeCommand(cmd)
		conn.Write([]byte(response))
	}

}

func (s *Server) executeCommand(cmd *protocol.Command) string {

	s.logger.WithFields(logrus.Fields{
		"cmd":  cmd.Name,
		"args": cmd.Args,
	}).Debug("Command received")

	switch cmd.Name {
	case "SET":
		if len(cmd.Args) != 2 {
			err := "wrong number of arguments for 'set', expected 2"
			s.logger.WithField("error", err).Warn("Command failed")
			return protocol.Error("wrong number of arguments for 'set', expected 2")
		}

		key, value := cmd.Args[0], cmd.Args[1]
		s.storage.Set(key, value)

		s.logger.WithFields(logrus.Fields{
			"key":   key,
			"event": "set",
		}).Info("Key set")

		return protocol.OK()

	case "GET":
		if len(cmd.Args) != 1 {
			err := "wrong number of arguments for 'get', expected 1"
			s.logger.WithField("error", err).Warn("Command failed")
			return protocol.Error("wrong number of arguments for 'get', expected 1")
		}

		key := cmd.Args[0]
		value, ok := s.storage.Get(key)

		status := "hit"
		if !ok {
			status = "miss"
		}

		s.logger.WithFields(logrus.Fields{
			"key":    key,
			"event":  "get",
			"status": status,
		}).Info("Key retrieved")

		if !ok {
			return protocol.NullBulk()
		}

		return protocol.BulkString(value.Data)

	case "DEL":
		if len(cmd.Args) != 1 {
			err := "wrong number of arguments for 'del', expected 1"
			s.logger.WithField("error", err).Warn("Command failed")
			return protocol.Error("wrong number of arguments for 'delete', expected 1")
		}

		key := cmd.Args[0]
		deleted := s.storage.Delete(key)

		event := "delete"
		status := "deleted"
		if !deleted {
			status = "not_found"
		}

		s.logger.WithFields(logrus.Fields{
			"key":    key,
			"event":  event,
			"status": status,
		}).Info("Key deletion attempted")

		if deleted {
			return protocol.WriteSimpleString("1")
		} else {
			return protocol.WriteSimpleString("0")
		}

	case "TTL":
		if len(cmd.Args) != 1 {
			err := "wrong number of arguments for 'ttl', expected 1"
			s.logger.WithField("error", err).Warn("Command failed")
			return protocol.Error("wrong number of arguments for 'ttl', expected 1")
		}

		key := cmd.Args[0]
		value, exists := s.storage.Get(key)

		if !exists {
			return protocol.WriteSimpleString("-2")
		}

		if value.ExpiresAt == nil {
			return protocol.WriteSimpleString("-1")
		}

		ttl := int64(time.Until(*value.ExpiresAt).Seconds())
		if ttl <= 0 {
			return protocol.WriteSimpleString("-2")
		}

		s.logger.WithFields(logrus.Fields{
			"key":    key,
			"ttl":    ttl,
			"event":  "ttl",
			"status": "valid",
		}).Info("TTL returned")

		return protocol.Integer(ttl)

	case "SETEX":
		if len(cmd.Args) != 3 {
			err := "wrong number of arguments for 'setex', expected 3"
			s.logger.WithField("error", err).Warn("Command failed")
			return protocol.Error("wrong number of arguments for 'setex', expected 3")
		}

		key := cmd.Args[0]
		duration, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return protocol.Error("invalid expire time")
		}

		value := cmd.Args[2]
		s.storage.SetEx(key, value, time.Duration(duration)*time.Second)

		s.logger.WithFields(logrus.Fields{
			"key":   key,
			"ttl":   duration,
			"event": "setex",
		}).Info("Key set with TTL")

		return protocol.OK()

	case "PING":
		if len(cmd.Args) == 0 {
			return protocol.WriteSimpleString("PONG")
		} else if len(cmd.Args) == 1 {
			return protocol.BulkString(cmd.Args[0])
		} else {
			return protocol.Error("wrong number of arguments for 'ping', expected 0 or 1")
		}

	default:
		s.logger.WithField("cmd", cmd.Name).Warn("Unknown command")
		return protocol.Error("unknown command '" + cmd.Name + "'")
	}
}
