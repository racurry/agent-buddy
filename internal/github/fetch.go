package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// gitCloneFunc is the function used to run git clone. Tests can override this.
var gitCloneFunc = runGitClone

// FetchAndExtract clones a GitHub repo and returns the path to a temp directory
// containing the repo contents. Tries SSH first, then falls back to HTTPS.
// The ref parameter can be a branch, tag, or commit SHA. If empty, uses the repo default branch.
func FetchAndExtract(orgRepo, ref string) (string, error) {
	parts := strings.SplitN(orgRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repo format %q, expected org/repo", orgRepo)
	}

	sshURL := fmt.Sprintf("git@github.com:%s.git", orgRepo)
	httpsURL := fmt.Sprintf("https://github.com/%s.git", orgRepo)

	// Try SSH first (uses existing SSH keys/config), then HTTPS
	for _, url := range []string{sshURL, httpsURL} {
		tmpDir, err := cloneRepo(url, ref)
		if err == nil {
			return tmpDir, nil
		}
	}

	return "", fmt.Errorf("could not clone %q via SSH or HTTPS", orgRepo)
}

func cloneRepo(repoURL, ref string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "agent-buddy-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	args := []string{"clone", "--depth", "1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repoURL, tmpDir)

	if err := gitCloneFunc(args); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone %s: %w", repoURL, err)
	}

	// Remove .git directory — we only need the working tree
	os.RemoveAll(tmpDir + "/.git")

	return tmpDir, nil
}

func runGitClone(args []string) error {
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
