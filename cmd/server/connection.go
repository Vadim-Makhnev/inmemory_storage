package main

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"redisclone/internal/protocol"
	"strings"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()

		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Error("read error", "err", err)
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
	switch cmd.Name {
	case "SET":
		if len(cmd.Args) != 2 {
			return protocol.Error("wrong number of arguments for 'set', expected 2")
		}

		key, value := cmd.Args[0], cmd.Args[1]
		s.storage.Set(key, value)

		return protocol.OK()

	case "GET":
		if len(cmd.Args) != 1 {
			return protocol.Error("wrong number of arguments for 'get', expected 1")
		}

		key := cmd.Args[0]
		value, ok := s.storage.Get(key)

		if !ok {
			return protocol.NullBulk()
		}

		return protocol.BulkString(value.Data)

	case "DEL":
		if len(cmd.Args) != 1 {
			return protocol.Error("wrong number of arguments for 'delete', expected 1")
		}

		key := cmd.Args[0]
		deleted := s.storage.Delete(key)

		if deleted {
			return protocol.WriteSimpleString("1")
		} else {
			return protocol.WriteSimpleString("0")
		}

	case "PING":
		if len(cmd.Args) == 0 {
			return protocol.WriteSimpleString("PONG")
		} else if len(cmd.Args) == 1 {
			return protocol.BulkString(cmd.Args[0])
		} else {
			return protocol.Error("wrong number of arguments for 'ping', expected 0 or 1")
		}

	default:
		return protocol.Error("unknown command '" + cmd.Name + "'")
	}
}
