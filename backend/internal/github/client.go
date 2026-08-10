package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GitHubRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

type installationReposResponse struct {
	TotalCount   int                `json:"total_count"`
	Repositories []GitHubRepository `json:"repositories"`
}

// ListInstallationRepos queries GET /installation/repositories using Installation Access Token
func ListInstallationRepos(installationToken string) ([]GitHubRepository, error) {
	if installationToken == "" {
		return nil, fmt.Errorf("installationToken is required")
	}

	url := "https://api.github.com/installation/repositories"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+installationToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list installation repos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github list repos API failed with status: %d", resp.StatusCode)
	}

	var result installationReposResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode repos response: %w", err)
	}
	return result.Repositories, nil
}
