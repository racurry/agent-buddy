package github

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchAndExtract_InvalidFormats(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"noslash"},
		{"/leading-slash"},
		{"trailing-slash/"},
		{"/"},
	}

	for _, tt := range tests {
		_, err := FetchAndExtract(tt.input, "main")
		if err == nil {
			t.Errorf("expected error for input %q, got nil", tt.input)
		}
	}
}

func TestFetchAndExtract_SSHSuccess(t *testing.T) {
	old := gitCloneFunc
	defer func() { gitCloneFunc = old }()

	var clonedURL string
	gitCloneFunc = func(args []string) error {
		// args: clone --depth 1 [--branch ref] <url> <dir>
		url := args[len(args)-2]
		dir := args[len(args)-1]
		clonedURL = url
		// Create a marker file so we can verify extraction
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("test"), 0644)
		return nil
	}

	tmpDir, err := FetchAndExtract("org/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if clonedURL != "git@github.com:org/repo.git" {
		t.Errorf("expected SSH URL, got %s", clonedURL)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "SKILL.md")); err != nil {
		t.Error("expected SKILL.md in extracted dir")
	}
}

func TestFetchAndExtract_SSHFailsFallsBackToHTTPS(t *testing.T) {
	old := gitCloneFunc
	defer func() { gitCloneFunc = old }()

	var urls []string
	gitCloneFunc = func(args []string) error {
		url := args[len(args)-2]
		dir := args[len(args)-1]
		urls = append(urls, url)
		if url == "git@github.com:org/repo.git" {
			return fmt.Errorf("ssh failed")
		}
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("test"), 0644)
		return nil
	}

	tmpDir, err := FetchAndExtract("org/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if len(urls) != 2 {
		t.Fatalf("expected 2 clone attempts, got %d", len(urls))
	}
	if urls[0] != "git@github.com:org/repo.git" {
		t.Errorf("first attempt should be SSH, got %s", urls[0])
	}
	if urls[1] != "https://github.com/org/repo.git" {
		t.Errorf("second attempt should be HTTPS, got %s", urls[1])
	}
}

func TestFetchAndExtract_BothFail(t *testing.T) {
	old := gitCloneFunc
	defer func() { gitCloneFunc = old }()

	gitCloneFunc = func(args []string) error {
		return fmt.Errorf("clone failed")
	}

	_, err := FetchAndExtract("org/repo", "main")
	if err == nil {
		t.Fatal("expected error when both SSH and HTTPS fail")
	}
	if got := err.Error(); got != `could not clone "org/repo" via SSH or HTTPS` {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestFetchAndExtract_NoRef(t *testing.T) {
	old := gitCloneFunc
	defer func() { gitCloneFunc = old }()

	var capturedArgs []string
	gitCloneFunc = func(args []string) error {
		capturedArgs = args
		dir := args[len(args)-1]
		os.MkdirAll(dir, 0755)
		return nil
	}

	tmpDir, err := FetchAndExtract("org/repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should be: clone --depth 1 <url> <dir> (no --branch flag)
	for _, arg := range capturedArgs {
		if arg == "--branch" {
			t.Error("--branch flag should not be present when ref is empty")
		}
	}
}

func TestFetchAndExtract_WithRef(t *testing.T) {
	old := gitCloneFunc
	defer func() { gitCloneFunc = old }()

	var capturedArgs []string
	gitCloneFunc = func(args []string) error {
		capturedArgs = args
		dir := args[len(args)-1]
		os.MkdirAll(dir, 0755)
		return nil
	}

	tmpDir, err := FetchAndExtract("org/repo", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	foundBranch := false
	for i, arg := range capturedArgs {
		if arg == "--branch" && i+1 < len(capturedArgs) && capturedArgs[i+1] == "v1.0.0" {
			foundBranch = true
		}
	}
	if !foundBranch {
		t.Errorf("expected --branch v1.0.0 in args: %v", capturedArgs)
	}
}
