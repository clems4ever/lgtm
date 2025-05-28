package github_test

import (
	"testing"

	"github.com/clems4ever/lgtm/internal/github"
)

func TestParsePullRequestURL(t *testing.T) {
	tests := []struct {
		input   string
		want    github.PRLink
		wantErr bool
	}{
		{
			input:   "https://github.com/owner/repo/pull/123",
			want:    github.PRLink{Owner: "owner", Repo: "repo", PRNumber: 123},
			wantErr: false,
		},
		{
			input:   "https://github.com/foo/bar/pull/1",
			want:    github.PRLink{Owner: "foo", Repo: "bar", PRNumber: 1},
			wantErr: false,
		},
		{
			input:   "https://github.com/foo/bar/pull/notanumber",
			wantErr: true,
		},
		{
			input:   "https://github.com/foo/bar/issues/123",
			wantErr: true,
		},
		{
			input:   "not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := github.ParsePullRequestURL(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParsePullRequestURL(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePullRequestURL(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParsePullRequestURL(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestPRLink_String(t *testing.T) {
	link := github.PRLink{Owner: "foo", Repo: "bar", PRNumber: 42}
	want := "https://github.com/foo/bar/pull/42"
	if got := link.String(); got != want {
		t.Errorf("PRLink.String() = %q, want %q", got, want)
	}
}

func TestPRLink_RepoFullName(t *testing.T) {
	link := github.PRLink{Owner: "foo", Repo: "bar", PRNumber: 42}
	want := "foo/bar"
	if got := link.RepoFullName(); got != want {
		t.Errorf("PRLink.RepoFullName() = %q, want %q", got, want)
	}
}
