package github

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetLatestReleaseAssets(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		jsonBody := `{
			"tag_name": "v9.0.0",
			"assets": [
				{"name": "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", "browser_download_url": "https://github.com/sharkdp/fd/releases/download/v9.0.0/fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz"},
				{"name": "fd-v9.0.0-aarch64-unknown-linux-gnu.tar.gz", "browser_download_url": "https://github.com/sharkdp/fd/releases/download/v9.0.0/fd-v9.0.0-aarch64-unknown-linux-gnu.tar.gz"}
			]
		}`

		httpGet = func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(jsonBody)),
			}, nil
		}

		s := &Client{}
		assets, tag, err := s.GetLatestReleaseAssets("sharkdp", "fd")

		assert.NoError(t, err)
		assert.Equal(t, "v9.0.0", tag)
		assert.Len(t, assets, 2)
		assert.Equal(t, "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", assets[0].Name)
		assert.Equal(t, "https://github.com/sharkdp/fd/releases/download/v9.0.0/fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", assets[0].DownloadURL)
	})

	t.Run("http error", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		httpGet = func(url string) (*http.Response, error) {
			return nil, fmt.Errorf("network error")
		}

		s := &Client{}
		_, _, err := s.GetLatestReleaseAssets("user", "repo")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("non-200 status", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		httpGet = func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}

		s := &Client{}
		_, _, err := s.GetLatestReleaseAssets("user", "nonexistent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("malformed JSON", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		httpGet = func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("{not json")),
			}, nil
		}

		s := &Client{}
		_, _, err := s.GetLatestReleaseAssets("user", "repo")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})

	t.Run("read body error", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		httpGet = func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(&errorReader{}),
			}, nil
		}

		s := &Client{}
		_, _, err := s.GetLatestReleaseAssets("user", "repo")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read response")
	})
}

func TestClient_GetReleaseAssetsByTag(t *testing.T) {
	t.Run("success with exact tag", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		jsonBody := `{
			"tag_name": "v9.0.0",
			"assets": [
				{"name": "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", "browser_download_url": "https://github.com/sharkdp/fd/releases/download/v9.0.0/fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz"}
			]
		}`

		httpGet = func(url string) (*http.Response, error) {
			assert.Contains(t, url, "/releases/tags/v9.0.0")
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(jsonBody)),
			}, nil
		}

		s := &Client{}
		assets, tag, err := s.GetReleaseAssetsByTag("sharkdp", "fd", "v9.0.0")

		assert.NoError(t, err)
		assert.Equal(t, "v9.0.0", tag)
		assert.Len(t, assets, 1)
	})

	t.Run("fallback adds v prefix", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		jsonBody := `{
			"tag_name": "v9.0.0",
			"assets": [
				{"name": "fd-v9.0.0-x86_64-unknown-linux-gnu.tar.gz", "browser_download_url": "https://example.com/fd.tar.gz"}
			]
		}`

		callCount := 0
		httpGet = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				assert.Contains(t, url, "/releases/tags/9.0.0")
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			assert.Contains(t, url, "/releases/tags/v9.0.0")
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(jsonBody)),
			}, nil
		}

		s := &Client{}
		assets, tag, err := s.GetReleaseAssetsByTag("sharkdp", "fd", "9.0.0")

		assert.NoError(t, err)
		assert.Equal(t, "v9.0.0", tag)
		assert.Len(t, assets, 1)
		assert.Equal(t, 2, callCount)
	})

	t.Run("no fallback when tag already has v prefix", func(t *testing.T) {
		origGet := httpGet
		defer func() { httpGet = origGet }()

		callCount := 0
		httpGet = func(url string) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}

		s := &Client{}
		_, _, err := s.GetReleaseAssetsByTag("sharkdp", "fd", "v99.0.0")

		assert.Error(t, err)
		assert.Equal(t, 1, callCount, "should not retry when tag already starts with v")
	})
}

type errorReader struct{}

func (e *errorReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
}

func TestNewService(t *testing.T) {
	svc := NewService()
	assert.NotNil(t, svc)
}
