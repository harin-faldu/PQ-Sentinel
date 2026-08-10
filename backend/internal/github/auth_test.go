package github

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAppJWT(t *testing.T) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test RSA key: %v", err)
	}

	appID := "123456"
	tokenStr, err := GenerateAppJWT(appID, privateKey)
	if err != nil {
		t.Fatalf("GenerateAppJWT failed: %v", err)
	}

	// Parse and verify JWT claims
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			t.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return &privateKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		t.Fatalf("Generated JWT token is invalid: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Failed to parse claims")
	}

	if claims["iss"] != appID {
		t.Errorf("Expected iss claim '%s', got '%v'", appID, claims["iss"])
	}

	exp := int64(claims["exp"].(float64))
	if exp <= time.Now().Unix() {
		t.Errorf("JWT token is already expired")
	}
}

func TestGetInstallationTokenValidation(t *testing.T) {
	_, err := GetInstallationToken("", 0)
	if err == nil {
		t.Errorf("Expected error for empty inputs, got nil")
	}
}
