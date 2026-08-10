package controllers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/quantumpq/qv-cbom/internal/github"
)

type GitHubController struct {
	appID      string
	privateKey *rsa.PrivateKey
}

func NewGitHubController(appID string, privateKey *rsa.PrivateKey) *GitHubController {
	if appID == "" {
		appID = os.Getenv("GITHUB_APP_ID")
	}
	if privateKey == nil {
		privateKey = loadPrivateKeyFromEnv()
	}
	return &GitHubController{
		appID:      appID,
		privateKey: privateKey,
	}
}

func (c *GitHubController) GetAppID() string {
	return c.appID
}

func (c *GitHubController) GetPrivateKey() *rsa.PrivateKey {
	return c.privateKey
}

func loadPrivateKeyFromEnv() *rsa.PrivateKey {
	pemStr := os.Getenv("GITHUB_PRIVATE_KEY")
	if pemStr != "" {
		key, _ := parseRSAPrivateKey([]byte(pemStr))
		if key != nil {
			return key
		}
	}

	keyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	if keyPath != "" {
		data, err := os.ReadFile(keyPath)
		if err == nil {
			key, _ := parseRSAPrivateKey(data)
			if key != nil {
				return key
			}
		}
	}
	return nil
}

func parseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := keyInterface.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	return nil, errors.New("unsupported RSA private key format")
}

// HandleListRepos handles GET /api/v1/github/repos?installation_id=12345
func (c *GitHubController) HandleListRepos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	installationIDStr := r.URL.Query().Get("installation_id")
	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil || installationID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid or missing installation_id query parameter",
		})
		return
	}

	if c.appID == "" || c.privateKey == nil {
		w.WriteHeader(http.StatusPreconditionFailed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("GitHub App private key is not configured on the backend server. To test GitHub App ingestion, please set GITHUB_APP_ID and GITHUB_PRIVATE_KEY (or GITHUB_PRIVATE_KEY_PATH) environment variables."),
		})
		return
	}

	appJWT, err := github.GenerateAppJWT(c.appID, c.privateKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to generate GitHub App RS256 JWT: %v", err),
		})
		return
	}

	tokenResp, err := github.GetInstallationToken(appJWT, installationID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("GitHub Installation Token exchange failed for ID %d: %v", installationID, err),
		})
		return
	}

	repos, err := github.ListInstallationRepos(tokenResp.Token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to list installation repositories: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(repos)
}

var ServerBootID = fmt.Sprintf("%d", time.Now().UnixNano())

// HandleInstallURL returns Vercel-style target selector URL with state payload
func (c *GitHubController) HandleInstallURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	appSlug := os.Getenv("GITHUB_APP_SLUG")
	if appSlug == "" {
		appSlug = "quantumvault-cbom-scanner"
	}

	statePayload := fmt.Sprintf(`{"from":"/new","timestamp":%d}`, time.Now().Unix())
	encodedState := base64.URLEncoding.EncodeToString([]byte(statePayload))

	selectTargetURL := fmt.Sprintf("https://github.com/apps/%s/installations/select_target?state=%s", appSlug, encodedState)
	directInstallURL := fmt.Sprintf("https://github.com/apps/%s/installations/new", appSlug)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_instance_id": ServerBootID,
		"install_url":        selectTargetURL,
		"select_target_url":  selectTargetURL,
		"direct_install_url": directInstallURL,
		"logout_url":         "https://github.com/logout",
		"app_slug":           appSlug,
		"app_configured":     c.appID != "" && c.privateKey != nil,
	})
}
