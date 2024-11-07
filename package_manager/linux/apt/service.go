package apt

// NewAptService to set up the apt service
func NewAptService() IApt {
	return &SApt{}
}
