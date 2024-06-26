// Package server implements the request/response logic of the server.
package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/airforce270/mc-srv/crypto"
	"github.com/airforce270/mc-srv/packet/config"
	"github.com/airforce270/mc-srv/packet/login"
	"github.com/airforce270/mc-srv/packet/readpacket"
	"github.com/airforce270/mc-srv/packet/slp"
	"github.com/airforce270/mc-srv/packet/types"
	"github.com/airforce270/mc-srv/server/keepaliver"
	"github.com/airforce270/mc-srv/server/serverstate"
	"github.com/google/uuid"
)

const (
	serverID = ""

	pingInterval      = 5 * time.Second
	keepAliveInterval = 5 * time.Second
)

// errEnableEncryption is not an error per se,
// but indicates to the caller that encryption should be used
// for all future reads and writes.
var errEnableEncryption = errors.New("enable encryption")

var mojangHasJoinedURL = url.URL{Scheme: "https", Host: "sessionserver.mojang.com", Path: "session/minecraft/hasJoined"}

type Conn struct {
	state  serverstate.State
	conn   net.Conn
	logger *log.Logger

	br *bufio.Reader
	bw *bufio.Writer

	playerUsername string
	playerUUID     uuid.UUID
	sharedSecret   []byte
	verifyToken    []byte
}

func NewConn(conn net.Conn) (*Conn, error) {
	verifyToken := make([]byte, 4)
	if _, err := crypto.RandReader.Read(verifyToken); err != nil {
		return nil, fmt.Errorf("failed to generate verify token: %w", err)
	}

	logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", conn.RemoteAddr().String()), log.Flags()|log.Lmsgprefix)

	return &Conn{
		state:       serverstate.PreHandshake,
		conn:        conn,
		logger:      logger,
		br:          newLoggingReader(conn, logger),
		bw:          newLoggingWriter(conn, logger),
		verifyToken: verifyToken,
	}, nil
}

// handleConn handles a new connection.
func (c *Conn) Handle(ctx context.Context) {
	var r io.Reader = c.br
	var w io.Writer = c.bw

	for {
		select {
		case <-ctx.Done():
			c.logger.Print("Context done, closing conn")
			return
		default:
		}

		err := c.handlePacket(ctx, r, w)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, crypto.ErrCloseConn) {
				c.logger.Printf("Failed to handle packet, closing conn: %v", err)
				return
			} else if errors.Is(err, errEnableEncryption) {
				c.logger.Printf("Enabling encryption for read stream...")
				if cr, err := crypto.NewDecryptReader(c.conn, c.sharedSecret); err == nil {
					r = newLoggingReader(cr, c.logger)
					c.logger.Printf("Enabled encryption for read stream.")
				} else {
					c.logger.Printf("Failed to enable encryption for read stream: %v", err)
				}

				c.logger.Printf("Enabling encryption for write stream...")
				if cw, err := crypto.NewEncryptWriter(w, c.sharedSecret); err == nil {
					w = newLoggingWriter(cw, c.logger)
					c.logger.Printf("Enabled encryption for write stream.")
				} else {
					c.logger.Printf("Failed to enable encryption for write stream: %v", err)
				}

			} else {
				c.logger.Printf("Failed to handle packet: %v", err)
			}
		}
		if err := c.bw.Flush(); err != nil {
			c.logger.Printf("Flushing conn write buffer failed: %v", err)
		}
	}
}

// Disconnect disconnects the client and closes the conn.
func (c *Conn) Disconnect(reason types.TextComponent) {
	switch {
	case c.state < serverstate.LoginComplete:
		break
	case c.state < serverstate.ConfigurationComplete:
		p := config.Disconnect{Reason: reason}
		if err := p.Write(c.bw); err != nil {
			c.logger.Printf("Disconnecting: failed to write disconnect packet: %v", err)
			break
		}
		if err := c.bw.Flush(); err != nil {
			c.logger.Printf("Disconnecting: failed to flush writer: %v", err)
			break
		}
	}

	if err := c.Close(); err != nil {
		c.logger.Printf("Disconnecting: failed to close conn: %v", err)
	}
}

// Close closes the conn.
func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) handlePacket(ctx context.Context, r io.Reader, w io.Writer) error {
	p, err := readpacket.Read(r, c.state, c.logger)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("got EOF, closing: %w %w", err, crypto.ErrCloseConn)
		}
		return fmt.Errorf("failed to read packet: %w", err)
	}
	if p == nil {
		return nil
	}
	c.logger.Printf("Received %T%+v", p, p)

	switch pp := p.(type) {
	case slp.StatusRequest:
		// do nothing
	case slp.Handshake:
		switch pp.NextState {
		case slp.HandshakeNextStateStatus:
			c.state = serverstate.ClientRequestingStatus
			sr, err := slp.NewStatusResponse(int(pp.ProtocolVersion))
			if err != nil {
				return fmt.Errorf("failed to create status response: %w", err)
			}
			if err := sr.Write(w); err != nil {
				return fmt.Errorf("failed to write status response: %w", err)
			}
			c.logger.Print("Wrote status response")
		case slp.HandshakeNextStateLogin:
			c.state = serverstate.ClientRequestingLogin
		}
	case slp.HandshakePingRequest:
		err := slp.HandshakePingResponse{Payload: pp.Payload}.Write(w)
		if err != nil {
			return fmt.Errorf("failed to write ping response: %w  ", err)
		}
		c.logger.Print("Wrote ping response")
	case login.LoginStart:
		c.playerUsername = pp.PlayerName
		c.playerUUID = pp.PlayerUUID

		er := login.EncryptionRequest{
			ServerID:          serverID,
			PublicKeyLength:   int32(len(crypto.PublicKeyPKIX)),
			PublicKey:         crypto.PublicKeyPKIX,
			VerifyTokenLength: int32(len(c.verifyToken)),
			VerifyToken:       c.verifyToken,
		}
		if err := er.Write(w); err != nil {
			return fmt.Errorf("failed to write encryption request: %w", err)
		}
		c.logger.Print("Wrote encryption request")
		c.state = serverstate.EncryptionRequested
	case login.EncryptionResponse:
		var err error
		c.sharedSecret, err = crypto.PrivateKey.Decrypt(crypto.RandReader, pp.SharedSecret, crypto.DecryptOpts)
		if err != nil {
			return fmt.Errorf("failed to decrypt shared secret: %w", err)
		}

		hash := sha1.New()
		hash.Write(stringToASCII(serverID))
		hash.Write(c.sharedSecret)
		hash.Write(crypto.PublicKeyPKIX)

		reqURL := mojangHasJoinedURL
		reqURL.RawQuery = url.Values{
			"username": {c.playerUsername},
			"serverId": {minecraftDigest(hash)},
		}.Encode()
		resp, err := http.Get(reqURL.String())
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

		ls := login.LoginSuccess{
			UUID:     c.playerUUID,
			Username: c.playerUsername,
		}

		cw, err := crypto.NewEncryptWriter(w, c.sharedSecret)
		if err != nil {
			return fmt.Errorf("failed to create encrypter: %w", err)
		}

		c.logger.Print("Warning: next logged write is the encrypted bytes, not the raw bytes.")

		if err := ls.Write(cw, c.logger); err != nil {
			return fmt.Errorf("failed to write login success: %w", err)
		}
		c.logger.Print("Wrote login success")
		c.state = serverstate.LoginCompletePendingAcknowledgement
		return errEnableEncryption
	case login.LoginAcknowledgement:
		c.state = serverstate.LoginComplete
		keepAlive := keepaliver.New(keepAliveInterval, w)
		go keepAlive.StartPinging(ctx, c.logger)
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-keepAlive.Notifier():
				c.Disconnect(types.TextComponent{
					Text: "Timeout - failed to respond to keepalive",
				})
				return
			}
		}()
	case config.ConfigClientInformation:
	case config.AcknowledgeFinishConfiguration:
		c.state = serverstate.ConfigurationComplete
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

func newLoggingReader(r io.Reader, logger *log.Logger) *bufio.Reader {
	return bufio.NewReader(io.TeeReader(r, readLogger{log: logger}))
}

func newLoggingWriter(w io.Writer, logger *log.Logger) *bufio.Writer {
	return bufio.NewWriter(io.MultiWriter(w, writeLogger{log: logger}))
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
