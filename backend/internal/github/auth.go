package github

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type InstallationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GenerateAppJWT creates a short-lived (9 min per GitHub rules) RS256 JWT signed with App private key.
func GenerateAppJWT(appID string, privateKey *rsa.PrivateKey) (string, error) {
	if appID == "" || privateKey == nil {
		return "", fmt.Errorf("appID and privateKey are required")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Add(-30 * time.Second).Unix(), // clock drift compensation
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": appID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GetInstallationToken exchanges the App JWT for a 1-hour Installation Access Token.
func GetInstallationToken(appJWT string, installationID int64) (*InstallationTokenResponse, error) {
	if appJWT == "" || installationID <= 0 {
		return nil, fmt.Errorf("invalid appJWT or installationID")
	}
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "QuantumVault-CBOM-Scanner")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Message          string `json:"message"`
			DocumentationURL string `json:"documentation_url"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		msg := errResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("github installation token exchange failed (status %d): %s", resp.StatusCode, msg)
	}

	var result InstallationTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode installation token response: %w", err)
	}
	return &result, nil
}
