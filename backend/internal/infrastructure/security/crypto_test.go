package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptor(t *testing.T) {
	encryptor, err := NewAESEncryptor()
	require.NoError(t, err)

	t.Run("Encrypt and Decrypt with Base64", func(t *testing.T) {
		plaintext := "123.456.789-00"

		ciphertext, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)

		// Verificar que es base64 válido
		_, err = base64.StdEncoding.DecodeString(ciphertext)
		assert.NoError(t, err, "Ciphertext should be valid base64")

		decrypted, err := encryptor.Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("Blind Index Generation", func(t *testing.T) {
		plaintext := "123.456.789-00"

		bidx1 := encryptor.GenerateBlindIndex(plaintext)
		assert.NotEmpty(t, bidx1)

		// Verificar que es base64 válido
		_, err := base64.StdEncoding.DecodeString(bidx1)
		assert.NoError(t, err, "Blind index should be valid base64")

		bidx2 := encryptor.GenerateBlindIndex(plaintext)
		assert.Equal(t, bidx1, bidx2, "Deterministic blind index for same plaintext")

		differentPlaintext := "987.654.321-00"
		bidx3 := encryptor.GenerateBlindIndex(differentPlaintext)
		assert.NotEqual(t, bidx1, bidx3, "Different blind index for different plaintext")
	})

	t.Run("Encryption is deterministic with Base64", func(t *testing.T) {
		plaintext := "123.456.789-00"

		c1, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)

		c2, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)

		// Base64 es determinístico, misma entrada = misma salida
		assert.Equal(t, c1, c2, "Base64 encoding should be deterministic")

		d1, err := encryptor.Decrypt(c1)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, d1)
	})

	t.Run("Invalid ciphertext", func(t *testing.T) {
		_, err := encryptor.Decrypt("not-base64!!!")
		assert.Error(t, err, "Should fail with invalid base64")
	})
}
