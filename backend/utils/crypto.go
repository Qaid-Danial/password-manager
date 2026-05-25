package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plaintext with AES-256-GCM using the provided 32-byte key.
//
// A fresh random 12-byte nonce is generated for every call using crypto/rand.
// Nonce reuse with the same key would catastrophically break GCM confidentiality,
// so we never reuse nonces — each ciphertext carries its own nonce prefix.
//
// Output format stored in the database: base64(nonce || ciphertext || gcm_tag)
// GCM provides both confidentiality and authenticity: any tampering causes
// decryption to fail with an authentication error.
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// crypto/rand — never math/rand — is mandatory for nonce generation
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal appends ciphertext + GCM tag after the nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Returns a generic error on any failure to avoid
// leaking information that could enable padding / oracle attacks.
func Decrypt(encoded string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("decryption failed")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("decryption failed")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("decryption failed")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("decryption failed")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// A failure here means wrong key OR tampered ciphertext
		return "", errors.New("decryption failed")
	}

	return string(plaintext), nil
}
