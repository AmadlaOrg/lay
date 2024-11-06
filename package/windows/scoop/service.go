package scoop

// NewScoopService to set up the dpkg service
func NewScoopService() IScoop {
	return &SScoop{}
}
