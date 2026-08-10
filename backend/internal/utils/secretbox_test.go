package utils

import (
	"bytes"
	"os"
	"testing"
)

func TestSecretBoxEncryptionRoundtrip(t *testing.T) {
	os.Setenv("MASTER_KEY", "12345678901234567890123456789012") // 32 bytes
	defer os.Unsetenv("MASTER_KEY")

	originalSecret := []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----")

	encrypted, err := EncryptSecret(originalSecret)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}

	if encrypted == "" {
		t.Fatalf("Expected non-empty base64 string")
	}

	decrypted, err := DecryptSecret(encrypted)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}

	if !bytes.Equal(originalSecret, decrypted) {
		t.Errorf("Decrypted secret does not match original secret")
	}
}
