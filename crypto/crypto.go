package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
)

var (
	RandReader = rand.Reader

	PrivateKey *rsa.PrivateKey

	// PublicKeyPKIX is the public key in PKIX, ASN.1 DER form.
	// It is a SubjectPublicKeyInfo structure.
	PublicKeyPKIX []byte

	DecryptOpts = &rsa.PKCS1v15DecryptOptions{}

	ErrCloseConn = errors.New("close connection")
)

func init() {
	var err error
	PrivateKey, PublicKeyPKIX, err = generateKeys(RandReader)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate encryption keys: %v", err))
	}
}

func generateKeys(rander io.Reader) (key *rsa.PrivateKey, pkix []byte, err error) {
	key, err = rsa.GenerateKey(rander, 1024)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA keypair: %w", err)
	}

	key.Precompute()

	if err := key.Validate(); err != nil {
		return nil, nil, fmt.Errorf("generated RSA keypair failed validation: %w", err)
	}

	pkix, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key to PKIX: %w", err)
	}

	return key, pkix, nil
}
