package github

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchAndExtract downloads a GitHub repo tarball and extracts it to a temp directory.
// Returns the path to the extracted repo root directory.
func FetchAndExtract(orgRepo string) (string, error) {
	parts := strings.SplitN(orgRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid repo format %q, expected org/repo", orgRepo)
	}

	url := fmt.Sprintf("https://github.com/%s/tarball/main", orgRepo)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetching repo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned status %d for %s", resp.StatusCode, url)
	}

	tmpDir, err := os.MkdirTemp("", "agent-buddy-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("decompressing: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var rootDir string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("reading tarball: %w", err)
		}

		// GitHub tarballs have a top-level directory like org-repo-sha/
		// Strip it to get relative paths.
		name := header.Name
		slashIdx := strings.Index(name, "/")
		if slashIdx < 0 {
			continue
		}
		if rootDir == "" {
			rootDir = name[:slashIdx]
		}
		relPath := name[slashIdx+1:]
		if relPath == "" {
			continue
		}

		target := filepath.Join(tmpDir, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("creating directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("creating parent directory: %w", err)
			}
			f, err := os.Create(target)
			if err != nil {
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("creating file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				os.RemoveAll(tmpDir)
				return "", fmt.Errorf("writing file: %w", err)
			}
			f.Close()
		}
	}

	return tmpDir, nil
}
