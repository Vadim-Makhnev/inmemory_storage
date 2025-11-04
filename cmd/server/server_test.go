package main

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_SetGet_Del(t *testing.T) {
	server := NewServer(":4004", nil)

	go func() {
		if err := server.Start(); err != nil {
			t.Log("Server start error:", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	addr := server.addr
	defer server.Stop()

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	sendCommand := func(cmd string) string {
		_, err := conn.Write([]byte(cmd + "\r\n"))
		require.NoError(t, err)

		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		return strings.TrimSpace(line)
	}

	resp := sendCommand("SET foo bar")
	assert.Equal(t, "+OK", resp)

	resp = sendCommand("GET foo")
	assert.Equal(t, "$3", resp)

	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "bar\r\n", line)

	resp = sendCommand("DEL foo")
	assert.Equal(t, "+1", resp)

	resp = sendCommand("GET foo")
	assert.Equal(t, "$-1", resp)
}

func TestServer_SetEx_TTL(t *testing.T) {
	server := NewServer(":4005", nil)

	go func() {
		if err := server.Start(); err != nil {
			t.Log("Server start error:", err)
		}
	}()

	time.Sleep(50 * time.Microsecond)

	addr := server.addr
	defer server.Stop()

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	reader := bufio.NewReader(conn)

	sendCommand := func(cmd string) string {
		_, err := conn.Write([]byte(cmd + "\r\n"))
		require.NoError(t, err)

		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		resp := strings.TrimSpace(line)
		return resp
	}

	resp := sendCommand("SETEX foo 2 bar")
	assert.Equal(t, "+OK", resp)

	resp = sendCommand("GET foo")
	assert.Equal(t, "$3", resp)

	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "bar\r\n", line)

	resp = sendCommand("TTL foo")
	require.True(t, strings.HasPrefix(resp, ":"))
	ttlSeconds, err := strconv.Atoi(strings.TrimPrefix(resp, ":"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, ttlSeconds, 1)
	assert.LessOrEqual(t, ttlSeconds, 2)

	time.Sleep(3 * time.Second)

	resp = sendCommand("GET mykey")
	assert.Equal(t, "$-1", resp)

	resp = sendCommand("TTL mykey")
	assert.Equal(t, ":-2", resp)
}
