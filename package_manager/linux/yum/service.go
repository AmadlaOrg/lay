package yum

// NewYumService to set up the dpkg service
func NewYumService() IYum {
	return &SYum{}
}
