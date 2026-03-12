package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

var (
	// ErrInvalidCiphertext indicates the ciphertext is malformed
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Manager handles encryption/decryption for data at rest
type Manager struct {
	key []byte
}

// NewManager creates a new encryption manager.
// masterKey and salt must be provided — never hardcode these values.
func NewManager(masterKey, salt string) *Manager {
	// Derive a proper 256-bit key using PBKDF2
	key := pbkdf2.Key([]byte(masterKey), []byte(salt), 100000, 32, sha256.New)
	return &Manager{key: key}
}

// NewManagerFromConfig creates a manager using explicit key and salt values
// (typically sourced from config.Config).
func NewManagerFromConfig(key, salt string) (*Manager, error) {
	if key == "" {
		return nil, errors.New("encryption key must not be empty")
	}
	if salt == "" {
		return nil, errors.New("encryption salt must not be empty")
	}
	return NewManager(key, salt), nil
}

// NewManagerFromEnv creates a manager using ENCRYPTION_KEY and ENCRYPTION_SALT env vars.
func NewManagerFromEnv() (*Manager, error) {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		return nil, errors.New("ENCRYPTION_KEY environment variable not set")
	}
	salt := os.Getenv("ENCRYPTION_SALT")
	if salt == "" {
		return nil, errors.New("ENCRYPTION_SALT environment variable not set")
	}
	return NewManager(key, salt), nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (m *Manager) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext encrypted with Encrypt
func (m *Manager) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts raw bytes
func (m *Manager) EncryptBytes(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptBytes decrypts raw bytes
func (m *Manager) DecryptBytes(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertextBytes, nil)
}

// HashForLookup creates a deterministic hash for encrypted field lookups
// WARNING: This reveals equality, use sparingly
func (m *Manager) HashForLookup(value string) string {
	hash := sha256.Sum256(append(m.key, []byte(value)...))
	return base64.StdEncoding.EncodeToString(hash[:])
}
