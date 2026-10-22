package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

const redisAddr = "127.0.0.1:6379"

const pingReq = "*1\r\n$4\r\nPING\r\n"
const pingRes = "+PONG\r\n"

func connectRedis(t *testing.T) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", redisAddr, 2*time.Second)
	if err != nil {
		t.Fatalf("could not connect to Redis at %s: %v; is the server running?",
			redisAddr,
			err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return conn
}

func sendPing(t *testing.T, conn net.Conn) string {
	t.Helper()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("could not set connection deadline: %v", err)
	}

	_, err := fmt.Fprintf(conn, pingReq)
	if err != nil {
		t.Fatalf("could not send PING: %v", err)
	}

	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	return reply
}

func TestPing(t *testing.T) {
	conn := connectRedis(t)

	got := sendPing(t, conn)
	want := pingRes

	if got != want {
		t.Errorf("PING response = %q; want %q", got, want)
	}
}

func TestConcurrentClients(t *testing.T) {
	client1 := connectRedis(t)
	client2 := connectRedis(t)

	type result struct {
		client string
		reply  string
		err    error
	}

	start := make(chan struct{})
	results := make(chan result, 2)

	worker := func(name string, conn net.Conn) {
		<-start

		if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
			results <- result{client: name, err: err}
			return
		}

		_, err := fmt.Fprint(conn, pingReq)
		if err != nil {
			results <- result{client: name, err: err}
			return
		}

		reply, err := bufio.NewReader(conn).ReadString('\n')

		results <- result{
			client: name,
			reply:  reply,
			err:    err,
		}
	}

	go worker("client 1", client1)
	go worker("client 2", client2)

	close(start)

	for i := 0; i < 2; i++ {
		got := <-results

		if got.err != nil {
			t.Errorf("%s failed: %v", got.client, got.err)
			continue
		}

		if got.reply != pingRes {
			t.Errorf(
				"%s received %q; want %q",
				got.client,
				got.reply,
				pingRes,
			)
		}
	}
}
