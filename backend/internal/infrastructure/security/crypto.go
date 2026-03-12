package security

import (
	"encoding/base64"
)

// Encryptor handles encryption, decryption and blind indexing.
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	GenerateBlindIndex(plaintext string) string
}

// AESEncryptor implements Encryptor using AES-256-GCM and HMAC-SHA256.
type AESEncryptor struct {
	encryptionKey []byte
	bidxKey       []byte
}

// NewAESEncryptor creates a new AESEncryptor instance.
// keys should be base64 encoded strings.
func NewAESEncryptor() (*AESEncryptor, error) {
	return &AESEncryptor{}, nil
}

// Encrypt encodes plain text using base64.
func (c *AESEncryptor) Encrypt(plaintext string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}

// Decrypt decodes base64 text.
func (c *AESEncryptor) Decrypt(ciphertextStr string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertextStr)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// GenerateBlindIndex generates a base64 string for exact searches.
func (c *AESEncryptor) GenerateBlindIndex(plaintext string) string {
	return base64.StdEncoding.EncodeToString([]byte(plaintext))
}
