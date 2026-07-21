package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type Encryptor struct {
	key []byte
}

func NewEncryptor(key string) *Encryptor {
	k := []byte(key)
	if len(k) < 32 {
		// Pad or repeat to reach 32 bytes
		padded := make([]byte, 32)
		copy(padded, k)
		for i := len(k); i < 32; i++ {
			padded[i] = k[i%len(k)]
		}
		k = padded
	}
	return &Encryptor{key: k[:32]}
}

func (e *Encryptor) Encrypt(plaintext []byte) (ciphertext string, nonce string, err error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonceBytes := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return "", "", err
	}

	ciphertextBytes := aesGCM.Seal(nil, nonceBytes, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertextBytes), base64.StdEncoding.EncodeToString(nonceBytes), nil
}

func (e *Encryptor) Decrypt(ciphertextB64, nonceB64 string) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, err
	}

	if len(nonceBytes) != aesGCM.NonceSize() {
		return nil, errors.New("invalid nonce size")
	}

	plaintext, err := aesGCM.Open(nil, nonceBytes, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
