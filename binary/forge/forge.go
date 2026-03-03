package forge

// Asset represents a downloadable release asset (shared across forges)
type Asset struct {
	Name        string
	DownloadURL string
}

// Forge defines operations for fetching release assets from a code forge
type Forge interface {
	GetLatestReleaseAssets(owner, repo string) ([]Asset, string, error)
	GetReleaseAssetsByTag(owner, repo, tag string) ([]Asset, string, error)
}
