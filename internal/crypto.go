// Berisi logika enkripsi dan dektipsi
package internal

import (
	"crypto/cipher"
)

type Crypto struct {
	aead cipher.AEAD
}

func LoadKey(path string) ([]byte, error)

func NewCrypto(key []byte) (*Crypto, error)

func GenerateNonce(size int) ([]byte, error) // Pengecekan awal, mungkin nanti diubah

func (c *Crypto) Encrypt(
	plaintext []byte,
	aad []byte,
) (
	nonce []byte,
	ciphertext []byte,
	err error,
)

func (c *Crypto) Decrypt(
	nonce []byte,
	ciphertext []byte,
	aad []byte,
) ([]byte, error)
