package types

import "errors"

// ErrNotSupported indicates the operation is not supported by this package manager
var ErrNotSupported = errors.New("operation not supported by this package manager")

// SearchResult represents a standardized package search result
type SearchResult struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
}

// PackageInfo represents an installed package
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
