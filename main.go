package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/airforce270/mc-srv/packet"
)

var (
	portFlag = flag.Int("port", 25565, "Port to listen on.")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	addrStr := fmt.Sprintf("127.0.0.1:%d", *portFlag)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		log.Fatalf("Failed to resolve TCP address %s: %v", addrStr, err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *portFlag, err)
	}
	defer listener.Close()
	log.Printf("Listening on port %d", *portFlag)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalf("Failed to get next connection on listener: %v", err)
		}

		go func() {
			defer conn.Close()
			for {
				p, err := packet.Read(conn)
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Print("Got EOF, closing...")
						break
					}
					log.Fatalf("Failed to read packet: %v", err)
				}
				if p == nil {
					continue
				}
				log.Printf("Received %s packet: %+v", p.Name(), p)

				switch pp := p.(type) {
				case packet.StatusRequest:
					// do nothing
				case packet.Handshake:
					err := packet.WriteStatusResponse(conn, int(pp.ProtocolVersion))
					if err != nil {
						log.Fatalf("Failed to write status response: %v", err)
					} else {
						log.Print("Wrote status response")
					}
				case packet.Ping:
					err := packet.WritePingResponse(conn, pp.Payload)
					if err != nil {
						log.Fatalf("Failed to write ping response: %v", err)
					} else {
						log.Print("Wrote ping response")
					}
				}
			}
		}()
	}
}
