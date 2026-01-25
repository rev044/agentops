package ratchet

import (
	"testing"
)

func TestValidateArtifactPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"absolute unix path", "/home/user/workspace/research.md", false},
		{"absolute root path", "/tmp/file.txt", false},
		{"relative dot path", "./research.md", true},
		{"relative parent path", "../research.md", true},
		{"tilde path", "~/gt/research.md", true},
		{"no leading slash", "gt/research.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArtifactPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArtifactPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestExtractArtifactPaths(t *testing.T) {
	tests := []struct {
		name        string
		closeReason string
		wantPaths   []string
	}{
		{
			name:        "empty",
			closeReason: "",
			wantPaths:   nil,
		},
		{
			name:        "artifact with absolute path",
			closeReason: "Complete. Artifact: /home/user/workspace/research.md",
			wantPaths:   []string{"/home/user/workspace/research.md"},
		},
		{
			name:        "see with absolute path",
			closeReason: "Fixed. See /path/to/file.go:123",
			wantPaths:   []string{"/path/to/file.go:123"},
		},
		{
			name:        "multiple paths",
			closeReason: "Done. Artifact: /path/one.md See /path/two.go",
			wantPaths:   []string{"/path/one.md", "/path/two.go"},
		},
		{
			name:        "case insensitive",
			closeReason: "ARTIFACT: /upper/case.md artifact: /lower/case.md",
			wantPaths:   []string{"/upper/case.md", "/lower/case.md"},
		},
		{
			name:        "no path keywords",
			closeReason: "Done, tests passing",
			wantPaths:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractArtifactPaths(tt.closeReason)
			if len(got) != len(tt.wantPaths) {
				t.Errorf("ExtractArtifactPaths() = %v, want %v", got, tt.wantPaths)
				return
			}
			for i, path := range got {
				if path != tt.wantPaths[i] {
					t.Errorf("ExtractArtifactPaths()[%d] = %q, want %q", i, path, tt.wantPaths[i])
				}
			}
		})
	}
}

func TestValidateCloseReason(t *testing.T) {
	tests := []struct {
		name        string
		closeReason string
		wantIssues  int
	}{
		{
			name:        "valid absolute path",
			closeReason: "Complete. Artifact: /home/user/workspace/research.md",
			wantIssues:  0,
		},
		{
			name:        "no paths referenced",
			closeReason: "Done, tests passing",
			wantIssues:  0,
		},
		{
			name:        "relative path with dot",
			closeReason: "Artifact: ./research.md",
			wantIssues:  1,
		},
		{
			name:        "tilde path",
			closeReason: "Artifact: ~/gt/file.md",
			wantIssues:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := ValidateCloseReason(tt.closeReason)
			if len(issues) != tt.wantIssues {
				t.Errorf("ValidateCloseReason() found %d issues, want %d: %v", len(issues), tt.wantIssues, issues)
			}
		})
	}
}
