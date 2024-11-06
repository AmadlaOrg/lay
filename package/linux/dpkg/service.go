package dpkg

// NewDpkgService to set up the dpkg service
func NewDpkgService() IDpkg {
	return &SDpkg{}
}
