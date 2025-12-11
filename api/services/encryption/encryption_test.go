package encryption_test

import (
	"callsign/services/encryption"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	mgr := encryption.NewManager("test-encryption-key-32bytes!!")

	// Test basic encrypt/decrypt
	plaintext := "Hello, World! This is a secret message."
	encrypted, err := mgr.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := mgr.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptEmpty(t *testing.T) {
	mgr := encryption.NewManager("test-key")

	// Empty string should return empty
	encrypted, err := mgr.Encrypt("")
	assert.NoError(t, err)
	assert.Empty(t, encrypted)

	decrypted, err := mgr.Decrypt("")
	assert.NoError(t, err)
	assert.Empty(t, decrypted)
}

func TestDecryptInvalid(t *testing.T) {
	mgr := encryption.NewManager("test-key")

	// Invalid base64
	_, err := mgr.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)

	// Valid base64 but invalid ciphertext
	_, err = mgr.Decrypt("aGVsbG8=") // "hello" in base64
	assert.Error(t, err)
}

func TestEncryptBytes(t *testing.T) {
	mgr := encryption.NewManager("test-key")

	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	encrypted, err := mgr.EncryptBytes(data)
	require.NoError(t, err)
	assert.NotEqual(t, data, encrypted)

	decrypted, err := mgr.DecryptBytes(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestHashForLookup(t *testing.T) {
	mgr := encryption.NewManager("test-key")

	// Same input should always produce same hash
	hash1 := mgr.HashForLookup("test-value")
	hash2 := mgr.HashForLookup("test-value")
	assert.Equal(t, hash1, hash2)

	// Different input should produce different hash
	hash3 := mgr.HashForLookup("different-value")
	assert.NotEqual(t, hash1, hash3)
}

func TestDifferentKeysProduceDifferentCiphertext(t *testing.T) {
	mgr1 := encryption.NewManager("key-one")
	mgr2 := encryption.NewManager("key-two")

	plaintext := "same message"

	enc1, _ := mgr1.Encrypt(plaintext)
	enc2, _ := mgr2.Encrypt(plaintext)

	// Different keys should produce different ciphertext
	assert.NotEqual(t, enc1, enc2)

	// Cannot decrypt with wrong key
	_, err := mgr2.Decrypt(enc1)
	assert.Error(t, err)
}
