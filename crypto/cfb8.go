package crypto

// Adapted from: https://github.com/GoLangMc/minecraft-server/blob/ca992e94ba13b1e44ec8251fad59b1857e2eb980/impl/conn/crypto/cfb8.go

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

// NewEncryptWriter wraps an io.Writer with encryption using the given secret.
func NewEncryptWriter(w io.Writer, secret []byte) (io.Writer, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return cipher.StreamWriter{}, fmt.Errorf("failed to create AES cipher for encryption: %w", err)
	}

	e, err := newEncrypt(block, secret)
	if err != nil {
		return cipher.StreamWriter{}, fmt.Errorf("failed to create encrypter: %w", err)
	}

	return cipher.StreamWriter{W: w, S: e}, nil
}

// NewDecrypt wraps an io.Reader with decryption using the given secret.
func NewDecryptReader(r io.Reader, secret []byte) (io.Reader, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return cipher.StreamReader{}, fmt.Errorf("failed to create AES cipher for decryption: %w", err)
	}

	d, err := newDecrypt(block, secret)
	if err != nil {
		return cipher.StreamReader{}, fmt.Errorf("failed to create decrypter: %w", err)
	}

	return cipher.StreamReader{R: r, S: d}, nil
}

// NewEncryptAndDecrypt creates encryption and decryption streams
// for a given secret.
func NewEncryptAndDecrypt(secret []byte) (encrypt cipher.Stream, decrypt cipher.Stream, err error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	encrypt, err = newEncrypt(block, secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create encrypter: %w", err)
	}
	decrypt, err = newDecrypt(block, secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create decrypter: %w", err)
	}

	return encrypt, decrypt, nil
}

// NewEncrypt creates a new stream for encryption.
func NewEncrypt(secret []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher for encryption: %w", err)
	}
	return newEncrypt(block, secret)
}

// NewDecrypt creates a new stream for decryption.
func NewDecrypt(secret []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher for decryption: %w", err)
	}
	return newDecrypt(block, secret)
}

func newEncrypt(block cipher.Block, iv []byte) (cipher.Stream, error) {
	c, err := newCFB8(block, iv)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newDecrypt(block cipher.Block, iv []byte) (cipher.Stream, error) {
	c, err := newCFB8(block, iv)
	if err != nil {
		return nil, err
	}
	c.decrypt = true
	return c, nil
}

// newCFB8 creates a new cfb8.
//
// block should be an AES cipher, see here for what iv should be:
// https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Initialization_vector_(IV)
func newCFB8(block cipher.Block, iv []byte) (*cfb8, error) {
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		return nil, fmt.Errorf("iv length (%d) must equal block size (%d)", len(iv), block.BlockSize())
	}

	x := &cfb8{
		b:     block,
		sr:    make([]byte, blockSize*4),
		srEnc: make([]byte, blockSize),
	}

	copy(x.sr, iv)

	return x, nil
}

// cfb8 implements a cipher feedback encryption stream
// with a feedback size of 8.
//
// https://en.wikipedia.org/wiki/Block_cipher_mode_of_operation#Cipher_feedback_(CFB)
//
// This is not thread safe.
type cfb8 struct {
	b       cipher.Block
	sr      []byte
	srEnc   []byte
	srPos   int
	decrypt bool
}

// XORKeyStream performs encryption or decryption of the src data,
// placing the converted data into dst.
// This method satifies cipher.Stream.
func (x *cfb8) XORKeyStream(dst, src []byte) {
	blockSize := x.b.BlockSize()

	for i := 0; i < len(src); i++ {
		x.b.Encrypt(x.srEnc, x.sr[x.srPos:x.srPos+blockSize])

		var c byte
		if x.decrypt {
			c = src[i]
			dst[i] = c ^ x.srEnc[0]
		} else {
			c = src[i] ^ x.srEnc[0]
			dst[i] = c
		}

		x.sr[x.srPos+blockSize] = c
		x.srPos++

		if x.srPos+blockSize == len(x.sr) {
			copy(x.sr, x.sr[x.srPos:])
			x.srPos = 0
		}
	}
}
