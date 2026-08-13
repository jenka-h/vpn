// Berisi logika enkripsi dan dekripsi, memakai AES-Galois Counter Mode untuk 256.
package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
)

const KeySize = 32

type Crypto struct {
	aead cipher.AEAD
}

func LoadKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == KeySize {
		return data, nil
	}

	text := strings.TrimSpace(string(data))
	if decoded, err := hex.DecodeString(text); err == nil && len(decoded) == KeySize {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(text); err == nil && len(decoded) == KeySize {
		return decoded, nil
	}

	return nil, errors.New("key must be 32 raw bytes, 64 hex characters, or 44 base64 characters")
}

func NewCrypto(key []byte) (*Crypto, error) {
	if len(key) != KeySize {
		return nil, errors.New("AES-256-GCM requires a 32-byte key")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Crypto{aead: aead}, nil
}

func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, nonce)
	return nonce, err
}

func (c *Crypto) Encrypt(
	plaintext []byte,
	aad []byte,
) (
	nonce []byte,
	ciphertext []byte,
	err error,
) {
	nonce, err = GenerateNonce(c.aead.NonceSize())
	if err != nil {
		return nil, nil, err
	}
	ciphertext = c.aead.Seal(nil, nonce, plaintext, aad)
	return nonce, ciphertext, nil
}

func (c *Crypto) Decrypt(
	nonce []byte,
	ciphertext []byte,
	aad []byte,
) ([]byte, error) {
	return c.aead.Open(nil, nonce, ciphertext, aad)
}
