package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptor(t *testing.T) {
	// 32 byte key for AES-256 (base64 encoded)

	encryptor, err := NewAESEncryptor()
	require.NoError(t, err)

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		plaintext := "123.456.789-00"

		ciphertext, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := encryptor.Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("Blind Index Generation", func(t *testing.T) {
		plaintext := "123.456.789-00"

		bidx1 := encryptor.GenerateBlindIndex(plaintext)
		assert.NotEmpty(t, bidx1)

		bidx2 := encryptor.GenerateBlindIndex(plaintext)
		assert.Equal(t, bidx1, bidx2, "Deterministic blind index for same plaintext")

		differentPlaintext := "987.654.321-00"
		bidx3 := encryptor.GenerateBlindIndex(differentPlaintext)
		assert.NotEqual(t, bidx1, bidx3, "Different blind index for different plaintext")
	})

	t.Run("Encryption randomness (Nonce)", func(t *testing.T) {
		plaintext := "123.456.789-00"

		c1, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)

		c2, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)

		assert.NotEqual(t, c1, c2, "Ciphertexts should be different due to random nonces")

		d1, err := encryptor.Decrypt(c1)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, d1)

		d2, err := encryptor.Decrypt(c2)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, d2)
	})

	t.Run("Invalid ciphertext", func(t *testing.T) {
		_, err := encryptor.Decrypt("not-base64")
		assert.Error(t, err)

		_, err = encryptor.Decrypt(base64.StdEncoding.EncodeToString([]byte("too short")))
		assert.Error(t, err)
	})
}
