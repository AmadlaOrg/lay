package winget

// NewWingetService to set up the dpkg service
func NewWingetService() IWinget {
	return &SWinget{}
}
