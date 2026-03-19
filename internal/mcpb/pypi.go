package mcpb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// PyPIInfo holds the relevant fields from the PyPI JSON API.
type PyPIInfo struct {
	Name        string
	Version     string
	Summary     string
	Author      string
	AuthorEmail string
	License     string
	HomePageURL string
	// EntryModule and EntryFunc are parsed from console_scripts.
	EntryModule string
	EntryFunc   string
	// PythonRequires is the minimum Python version constraint.
	PythonRequires string
}

// pypiResponse mirrors the subset of the PyPI JSON API we need.
type pypiResponse struct {
	Info struct {
		Name           string `json:"name"`
		Version        string `json:"version"`
		Summary        string `json:"summary"`
		Author         string `json:"author"`
		AuthorEmail    string `json:"author_email"`
		License        string `json:"license"`
		HomePageURL    string `json:"home_page"`
		ProjectURL     string `json:"project_url"`
		RequiresPython string `json:"requires_python"`
	} `json:"info"`
	URLs []struct {
		PackageType string `json:"packagetype"`
		URL         string `json:"url"`
	} `json:"urls"`
}

// FetchPyPI fetches package metadata from the PyPI JSON API and extracts
// the console_scripts entry point.
func FetchPyPI(packageName string) (*PyPIInfo, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching PyPI metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("PyPI returned %d for package %q", resp.StatusCode, packageName)
	}

	var data pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding PyPI response: %w", err)
	}

	info := &PyPIInfo{
		Name:           data.Info.Name,
		Version:        data.Info.Version,
		Summary:        data.Info.Summary,
		Author:         data.Info.Author,
		AuthorEmail:    data.Info.AuthorEmail,
		License:        data.Info.License,
		HomePageURL:    data.Info.HomePageURL,
		PythonRequires: data.Info.RequiresPython,
	}

	if info.Author == "" && info.AuthorEmail != "" {
		// PyPI often has "Name <email>" in author_email; extract the name.
		if idx := strings.Index(info.AuthorEmail, "<"); idx > 0 {
			info.Author = strings.TrimSpace(info.AuthorEmail[:idx])
		} else {
			info.Author = info.AuthorEmail
		}
	}
	if info.Author == "" {
		info.Author = "Unknown"
	}
	if info.PythonRequires == "" {
		info.PythonRequires = ">=3.10"
	}

	// We need to find the entry point. PyPI JSON API doesn't directly expose
	// console_scripts, so we derive it from the package name convention:
	// package "foo-bar" typically has module "foo_bar" with entry point "main".
	// This matches the uvx convention where the script name == package name.
	module := strings.ReplaceAll(packageName, "-", "_")
	info.EntryModule = module
	info.EntryFunc = "main"

	return info, nil
}
