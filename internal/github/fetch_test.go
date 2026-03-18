package github

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func TestFetchAndExtract_HTTP404(t *testing.T) {
	oldClient := httpClient
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	defer func() { httpClient = oldClient }()

	_, err := FetchAndExtract("org/repo", "main")
	if err == nil || !strings.Contains(err.Error(), `not found: org/repo at ref "main"`) {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestFetchAndExtract_HTTP500(t *testing.T) {
	oldClient := httpClient
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("boom")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}
	defer func() { httpClient = oldClient }()

	_, err := FetchAndExtract("org/repo", "main")
	if err == nil || !strings.Contains(err.Error(), "GitHub returned status 500") {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

func TestFetchAndExtract_Timeout(t *testing.T) {
	oldClient := httpClient
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, &url.Error{Op: "Get", URL: req.URL.String(), Err: context.DeadlineExceeded}
	})}
	defer func() { httpClient = oldClient }()

	_, err := FetchAndExtract("org/repo", "main")
	if err == nil || !strings.Contains(err.Error(), "fetching repo") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
