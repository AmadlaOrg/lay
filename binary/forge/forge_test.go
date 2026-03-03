package forge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SourceInfo
		wantErr string
	}{
		{
			name:  "bare shorthand defaults to github",
			input: "sharkdp/fd",
			want:  SourceInfo{Forge: "github", Owner: "sharkdp", Repo: "fd", Version: "", IsForge: true},
		},
		{
			name:  "bare shorthand with version",
			input: "sharkdp/fd@v9.0.0",
			want:  SourceInfo{Forge: "github", Owner: "sharkdp", Repo: "fd", Version: "v9.0.0", IsForge: true},
		},
		{
			name:  "version without v prefix",
			input: "user/repo@1.2.3",
			want:  SourceInfo{Forge: "github", Owner: "user", Repo: "repo", Version: "1.2.3", IsForge: true},
		},
		{
			name:  "github prefix explicit",
			input: "github:sharkdp/fd",
			want:  SourceInfo{Forge: "github", Owner: "sharkdp", Repo: "fd", Version: "", IsForge: true},
		},
		{
			name:  "github prefix with version",
			input: "github:sharkdp/fd@v9.0.0",
			want:  SourceInfo{Forge: "github", Owner: "sharkdp", Repo: "fd", Version: "v9.0.0", IsForge: true},
		},
		{
			name:  "gitlab prefix",
			input: "gitlab:user/project",
			want:  SourceInfo{Forge: "gitlab", Owner: "user", Repo: "project", Version: "", IsForge: true},
		},
		{
			name:  "gitlab prefix with version",
			input: "gitlab:user/project@v2.0",
			want:  SourceInfo{Forge: "gitlab", Owner: "user", Repo: "project", Version: "v2.0", IsForge: true},
		},
		{
			name:  "codeberg prefix",
			input: "codeberg:user/repo",
			want:  SourceInfo{Forge: "codeberg", Owner: "user", Repo: "repo", Version: "", IsForge: true},
		},
		{
			name:  "codeberg prefix with version",
			input: "codeberg:user/repo@v1.0.0",
			want:  SourceInfo{Forge: "codeberg", Owner: "user", Repo: "repo", Version: "v1.0.0", IsForge: true},
		},
		{
			name:  "https URL is not forge",
			input: "https://example.com/tool.tar.gz",
			want:  SourceInfo{},
		},
		{
			name:  "http URL is not forge",
			input: "http://example.com/tool.tar.gz",
			want:  SourceInfo{},
		},
		{
			name:  "single word is not forge",
			input: "fd",
			want:  SourceInfo{},
		},
		{
			name:  "empty string",
			input: "",
			want:  SourceInfo{},
		},
		{
			name:    "unknown prefix returns error",
			input:   "bitbucket:user/repo",
			wantErr: "unsupported forge: bitbucket",
		},
		{
			name:  "hyphenated owner and repo",
			input: "my-org/my-repo",
			want:  SourceInfo{Forge: "github", Owner: "my-org", Repo: "my-repo", Version: "", IsForge: true},
		},
		{
			name:  "dotted repo name",
			input: "org/repo.go",
			want:  SourceInfo{Forge: "github", Owner: "org", Repo: "repo.go", Version: "", IsForge: true},
		},
		{
			name:    "empty remainder after prefix",
			input:   "unknown:",
			wantErr: "unsupported forge: unknown",
		},
		{
			name:    "leading space with colon",
			input:   "bogus:user/repo",
			wantErr: "unsupported forge: bogus",
		},
		{
			name:    "unknown prefix with colon in remainder",
			input:   "some:thing:else",
			wantErr: "unsupported forge: some",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.input)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Equal(t, SourceInfo{}, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
