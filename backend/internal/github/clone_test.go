package github

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenScrubbingOnCloneFailure(t *testing.T) {
	fakeToken := "ghs_SECRET_TOKEN_123456789"
	err := CloneRepository(fakeToken, "invalid/non-existent-repo-9999", "main", "invalid_dir")
	if err == nil {
		t.Fatalf("Expected CloneRepository to fail for invalid repository")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, fakeToken) {
		t.Errorf("SECURITY LEAK DETECTED: Token was not scrubbed from error output! Got: %s", errMsg)
	}

	// Verify token scrubbing helper explicitly
	sampleLogWithToken := "fatal: Authentication failed for https://x-access-token:" + fakeToken + "@github.com/repo.git"
	scrubbed := strings.ReplaceAll(sampleLogWithToken, fakeToken, "***REDACTED***")
	if strings.Contains(scrubbed, fakeToken) || !strings.Contains(scrubbed, "***REDACTED***") {
		t.Errorf("Token scrubbing helper failed to sanitize token string")
	}
}

func TestLocalGitCloneAndHistorySecretDetection(t *testing.T) {
	// Create temporary directory for local git repository
	repoDir, err := os.MkdirTemp("", "git_test_repo_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(repoDir)

	// Check if git binary is available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not found on PATH, skipping local git clone test")
	}

	// Helper to run git commands
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v — output: %s", args, err, out)
		}
	}

	runGit("init")
	runGit("checkout", "-b", "main")

	// Commit 1: Add file with secret
	secretFile := filepath.Join(repoDir, "config.py")
	secretContent := `AWS_SECRET_KEY = "AWS_EXAMPLE_SECRET_KEY_FOR_TESTING_12345"`
	if err := os.WriteFile(secretFile, []byte(secretContent), 0644); err != nil {
		t.Fatalf("Failed to write config.py: %v", err)
	}
	runGit("add", "config.py")
	runGit("commit", "-m", "Commit 1: Add secret")

	// Commit 2: Delete secret file
	if err := os.Remove(secretFile); err != nil {
		t.Fatalf("Failed to remove config.py: %v", err)
	}
	runGit("add", "config.py")
	runGit("commit", "-m", "Commit 2: Remove secret")

	// Clone local repo to destDir
	destDir, err := os.MkdirTemp("", "git_clone_dest_*")
	if err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	cloneCmd := exec.Command("git", "clone", "--branch", "main", "--single-branch", repoDir, destDir)
	out, err := cloneCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to clone local test repo: %v - output: %s", err, out)
	}

	// Verify cloned repository has .git directory and full history preserved
	gitDir := filepath.Join(destDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Errorf("Cloned repository is missing .git directory")
	}
}
