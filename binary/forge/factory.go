package forge

import (
	"fmt"
	"regexp"
	"strings"
)

var shorthandRegex = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)(?:@([a-zA-Z0-9_.-]+))?$`)

// SourceInfo holds parsed shorthand information
type SourceInfo struct {
	Forge   string // "github", "gitlab", "codeberg", or ""
	Owner   string
	Repo    string
	Version string
	IsForge bool // true if shorthand, false if direct URL
}

// ParseSource parses input into forge info.
//
//	"owner/repo"              -> github, owner, repo, "", true
//	"owner/repo@v1.0"         -> github, owner, repo, "v1.0", true
//	"github:owner/repo@v1.0"  -> github, owner, repo, "v1.0", true
//	"gitlab:owner/repo"       -> gitlab, owner, repo, "", true
//	"codeberg:owner/repo"     -> codeberg, owner, repo, "", true
//	"https://..."             -> "", "", "", "", false
func ParseSource(input string) (SourceInfo, error) {
	prefix := ""
	remainder := input

	// Check for forge prefix (e.g. "gitlab:owner/repo"), but not URLs ("https://...")
	if idx := strings.Index(input, ":"); idx > 0 && !strings.Contains(input, "://") {
		candidate := input[:idx]
		switch candidate {
		case "github", "gitlab", "codeberg":
			prefix = candidate
			remainder = input[idx+1:]
		default:
			return SourceInfo{}, fmt.Errorf("unsupported forge: %s", candidate)
		}
	}

	matches := shorthandRegex.FindStringSubmatch(remainder)
	if matches == nil {
		return SourceInfo{}, nil
	}

	forge := prefix
	if forge == "" {
		forge = "github"
	}

	return SourceInfo{
		Forge:   forge,
		Owner:   matches[1],
		Repo:    matches[2],
		Version: matches[3],
		IsForge: true,
	}, nil
}

