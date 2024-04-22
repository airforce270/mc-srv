// Package server implements the request/response logic of the server.
package server

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/airforce270/mc-srv/crypto"
	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/server/serverstate"
	"github.com/google/uuid"
)

const serverID = ""

type Conn struct {
	state serverstate.State

	playerUsername string
	playerUUID     uuid.UUID
	sharedSecret   []byte
	verifyToken    []byte
}

func NewConn() (*Conn, error) {
	verifyToken := make([]byte, 4)
	if _, err := crypto.RandReader.Read(verifyToken); err != nil {
		return nil, fmt.Errorf("failed to generate verify token: %w", err)
	}

	return &Conn{
		state:       serverstate.PreHandshake,
		verifyToken: verifyToken,
	}, nil
}

// handleConn handles a new connection.
func (c *Conn) Handle(conn net.Conn) {
	logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", conn.RemoteAddr().String()), log.Flags()|log.Lmsgprefix)
	r := bufio.NewReader(io.TeeReader(conn, readLogger{log: logger}))
	w := bufio.NewWriter(io.MultiWriter(conn, writeLogger{log: logger}))

	defer conn.Close()
	for {
		err := c.handlePacket(r, w, logger)
		if err != nil {
			if errors.Is(err, crypto.ErrCloseConn) {
				logger.Printf("Failed to handle packet, closing conn: %v", err)
				break
			} else {
				logger.Printf("Failed to handle packet: %v", err)
			}
		}
		if err := w.Flush(); err != nil {
			logger.Printf("Flushing conn write buffer failed: %v", err)
		}
	}
}

func (c *Conn) handlePacket(r io.Reader, w io.Writer, logger *log.Logger) error {
	p, err := packet.Read(r, c.state, logger)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("got EOF, closing: %w %w", err, crypto.ErrCloseConn)
		}
		return fmt.Errorf("failed to read packet: %w", err)
	}
	if p == nil {
		return nil
	}
	logger.Printf("Received %s: %+v", p.Name(), p)

	switch pp := p.(type) {
	case packet.StatusRequest:
		// do nothing
	case packet.Handshake:
		switch pp.NextState {
		case packet.HandshakeNextStateStatus:
			c.state = serverstate.ClientRequestingStatus
			sr, err := packet.NewStatusResponse(int(pp.ProtocolVersion))
			if err != nil {
				return fmt.Errorf("failed to create status response: %w", err)
			} else {
				if err := sr.Write(w); err != nil {
					return fmt.Errorf("failed to write status response: %w", err)
				}
				logger.Print("Wrote status response")
			}
		case packet.HandshakeNextStateLogin:
			c.state = serverstate.ClientRequestingLogin
		}
	case packet.PingRequest:
		err := packet.PingResponse{Payload: pp.Payload}.Write(w)
		if err != nil {
			return fmt.Errorf("failed to write ping response: %w  ", err)
		} else {
			logger.Print("Wrote ping response")
		}
	case packet.LoginStart:
		c.playerUsername = pp.PlayerName
		c.playerUUID = pp.PlayerUUID

		er := packet.EncryptionRequest{
			ServerID:          serverID,
			PublicKeyLength:   int32(len(crypto.PublicKeyPKIX)),
			PublicKey:         crypto.PublicKeyPKIX,
			VerifyTokenLength: int32(len(c.verifyToken)),
			VerifyToken:       c.verifyToken,
		}
		if err := er.Write(w); err != nil {
			return fmt.Errorf("failed to write encryption request: %w", err)
		} else {
			logger.Print("Wrote encryption request")
			c.state = serverstate.EncryptionRequested
		}
	case packet.EncryptionResponse:
		var err error
		c.sharedSecret, err = crypto.PrivateKey.Decrypt(crypto.RandReader, pp.SharedSecret, crypto.DecryptOpts)
		if err != nil {
			return fmt.Errorf("failed to decrypt shared secret: %w", err)
		}

		hash := sha1.New()
		hash.Write(stringToASCII(serverID))
		hash.Write(c.sharedSecret)
		hash.Write(crypto.PublicKeyPKIX)

		digest := minecraftDigest(hash)
		resp, err := http.Get(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s", c.playerUsername, digest))
		if err != nil {
			return fmt.Errorf("failed to call Mojang Session hasJoined API: %w", err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body from hasJoined call: %w", err)
		}
		var hasJoinedResp HasJoinedResponse
		if err := json.Unmarshal(body, &hasJoinedResp); err != nil {
			return fmt.Errorf("failed to unmarshal body from hasJoined call to JSON: %w", nil)
		}

		newPlayerUUID, err := uuid.Parse(hasJoinedResp.ID)
		if err != nil {
			return fmt.Errorf("failed to parse player's UUID (%s): %w", hasJoinedResp.ID, err)
		}
		if newPlayerUUID.String() != c.playerUUID.String() {
			return fmt.Errorf("new player UUID %s doesn't match the UUID we saw before: %s", newPlayerUUID, c.playerUUID)
		}
		if !strings.EqualFold(hasJoinedResp.Name, c.playerUsername) {
			return fmt.Errorf("new player username %s doesn't match the name we saw before: %s", hasJoinedResp.Name, c.playerUsername)
		}

		verifyToken, err := crypto.PrivateKey.Decrypt(crypto.RandReader, pp.VerifyToken, crypto.DecryptOpts)
		if err != nil {
			return fmt.Errorf("failed to decrypt verify token: %w", err)
		}
		if !bytes.Equal(verifyToken, c.verifyToken) {
			return fmt.Errorf("returned verify token (%x) does not match sent (%x), closing connection", verifyToken, c.verifyToken)
		}

		ls := packet.LoginSuccess{
			UUID:     c.playerUUID,
			Username: c.playerUsername,
		}

		cw, err := crypto.NewEncryptWriter(w, c.sharedSecret)
		if err != nil {
			return fmt.Errorf("failed to create encrypter: %w", err)
		}

		if err := ls.Write(cw, logger); err != nil {
			return fmt.Errorf("failed to write login success: %w", err)
		} else {
			logger.Print("Wrote login success")
			c.state = serverstate.LoginComplete
		}
	}

	return nil
}

// HasJoinedResponse is the response from the /hasJoined Mojang endpoint.
type HasJoinedResponse struct {
	// Player's identifier, in the format 11111111222233334444555555555555
	ID string `json:"id"`
	// Player's username
	Name string `json:"name"`
	// Other properties, normally has one containing the user's skin blob:
	// {Name: "textures", Value: "base64 string",
	//  Signature: "base64 string signed using Yggdrasil's private key'"}
	Properties []HasJoinedResponseProperty `json:"properties"`
}

// HasJoinedResponseProperty is a property in HasJoinedResponse.
type HasJoinedResponseProperty struct {
	// Name of the property.
	Name string `json:"name"`
	// Value of the property.
	Value string `json:"value"`
	// Signature of the value.
	Signature string `json:"signature"`
}

// stringToASCII converts a UTF-8 string to ASCII bytes.
func stringToASCII(s string) []byte {
	var b []byte
	for _, c := range s {
		if c > utf8.RuneSelf {
			c = '?'
		}
		b = append(b, byte(c))
	}
	return b
}
