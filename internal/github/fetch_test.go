package github

import (
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
		_, err := FetchAndExtract(tt.input)
		if err == nil {
			t.Errorf("expected error for input %q, got nil", tt.input)
		}
	}
}
