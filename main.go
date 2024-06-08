package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/airforce270/mc-srv/server"
)

var (
	portFlag = flag.Int("port", 25565, "Port to listen on.")
)

func createListener(port int) (*net.TCPListener, error) {
	addrStr := fmt.Sprintf("127.0.0.1:%d", port)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address %s: %v", addrStr, err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	return listener, err
}

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	listener, err := createListener(*portFlag)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()
	log.Printf("Listening on port %d", *portFlag)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		listener.Close()
	}()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalf("Failed to get next connection on listener: %v", err)
		}
		conn.SetNoDelay(true)
		conn.SetKeepAlive(true)
		log.Printf("New connection from %s", conn.RemoteAddr().String())

		c, err := server.NewConn(conn)
		if err != nil {
			log.Printf("Failed to create connection handler: %v", err)
			conn.Close()
			continue
		}
		defer c.Close()
		go func() {
			connCtx, cancelConn := context.WithCancel(ctx)
			c.Handle(connCtx)
			cancelConn()
		}()
	}
}
