package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func buildAuthenticatedCloneURL(installationToken, fullName string) string {
	return fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", installationToken, fullName)
}

// GetRemoteHeadSHA cheaply checks remote HEAD commit SHA without cloning
func GetRemoteHeadSHA(installationToken, fullName, branch string) (string, error) {
	if installationToken == "" || fullName == "" {
		return "", fmt.Errorf("installationToken and fullName are required")
	}
	if branch == "" {
		branch = "main"
	}
	url := buildAuthenticatedCloneURL(installationToken, fullName)
	cmd := exec.Command("git", "ls-remote", url, "refs/heads/"+branch)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git ls-remote failed: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("branch %q not found or repo inaccessible", branch)
	}
	return fields[0], nil
}

// CloneRepository performs a full (non-shallow), single-branch ephemeral clone into destDir.
// Full history is preserved so secret_scanner's `git log --all -p` can scan deleted secrets.
func CloneRepository(installationToken, fullName, branch, destDir string) error {
	if installationToken == "" || fullName == "" || destDir == "" {
		return fmt.Errorf("installationToken, fullName, and destDir are required")
	}
	if branch == "" {
		branch = "main"
	}
	url := buildAuthenticatedCloneURL(installationToken, fullName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone",
		"--branch", branch,
		"--single-branch",
		url, destDir,
	)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Scrub installation token from error output to prevent credential leaking
		scrubbed := strings.ReplaceAll(string(output), installationToken, "***REDACTED***")
		return fmt.Errorf("git clone failed: %w — output: %s", err, scrubbed)
	}
	return nil
}
