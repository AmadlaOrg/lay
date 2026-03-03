package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AmadlaOrg/lay/binary/forge"
)

// For mocking
var httpGet = http.Get

// release represents a GitHub release
type release struct {
	TagName string       `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

// releaseAsset represents a GitHub release asset
type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Client implements forge.Forge for GitHub
type Client struct{}

// GetLatestReleaseAssets fetches assets from the latest release of a GitHub repo.
func (s *Client) GetLatestReleaseAssets(owner, repo string) ([]forge.Asset, string, error) {
	return s.fetchRelease(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo))
}

// GetReleaseAssetsByTag fetches assets for a specific tagged release.
// If the tag doesn't have a "v" prefix and the request fails, it retries with a "v" prefix.
func (s *Client) GetReleaseAssetsByTag(owner, repo, tag string) ([]forge.Asset, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	assets, tagName, err := s.fetchRelease(url)
	if err != nil && len(tag) > 0 && tag[0] != 'v' {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/v%s", owner, repo, tag)
		return s.fetchRelease(url)
	}
	return assets, tagName, err
}

func (s *Client) fetchRelease(url string) ([]forge.Asset, string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	assets := make([]forge.Asset, len(rel.Assets))
	for i, a := range rel.Assets {
		assets[i] = forge.Asset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
		}
	}

	return assets, rel.TagName, nil
}
