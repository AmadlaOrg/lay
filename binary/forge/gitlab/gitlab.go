package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/AmadlaOrg/lay/binary/forge"
)

// For mocking
var httpGet = http.Get

type release struct {
	TagName string       `json:"tag_name"`
	Assets  releaseAssets `json:"assets"`
}

type releaseAssets struct {
	Links []releaseLink `json:"links"`
}

type releaseLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Client implements forge.Forge for GitLab
type Client struct{}

// GetLatestReleaseAssets fetches assets from the latest release of a GitLab project.
// GitLab has no /releases/latest endpoint, so we list releases and take the first.
func (s *Client) GetLatestReleaseAssets(owner, repo string) ([]forge.Asset, string, error) {
	project := url.PathEscape(owner + "/" + repo)
	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases", project)

	resp, err := httpGet(apiURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GitLab API returned status %d for %s", resp.StatusCode, apiURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	var releases []release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, "", fmt.Errorf("failed to parse releases JSON: %w", err)
	}

	if len(releases) == 0 {
		return nil, "", fmt.Errorf("no releases found for %s/%s", owner, repo)
	}

	rel := releases[0]
	return convertAssets(rel), rel.TagName, nil
}

// GetReleaseAssetsByTag fetches assets for a specific tagged release.
// If the tag doesn't have a "v" prefix and the request fails, it retries with a "v" prefix.
func (s *Client) GetReleaseAssetsByTag(owner, repo, tag string) ([]forge.Asset, string, error) {
	assets, tagName, err := s.fetchReleaseByTag(owner, repo, tag)
	if err != nil && len(tag) > 0 && tag[0] != 'v' {
		return s.fetchReleaseByTag(owner, repo, "v"+tag)
	}
	return assets, tagName, err
}

func (s *Client) fetchReleaseByTag(owner, repo, tag string) ([]forge.Asset, string, error) {
	project := url.PathEscape(owner + "/" + repo)
	apiURL := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases/%s", project, url.PathEscape(tag))

	resp, err := httpGet(apiURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GitLab API returned status %d for %s", resp.StatusCode, apiURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	return convertAssets(rel), rel.TagName, nil
}

func convertAssets(rel release) []forge.Asset {
	assets := make([]forge.Asset, len(rel.Assets.Links))
	for i, link := range rel.Assets.Links {
		assets[i] = forge.Asset{
			Name:        link.Name,
			DownloadURL: link.URL,
		}
	}
	return assets
}
