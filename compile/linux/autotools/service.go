package autotools

// NewAutotoolsService to set up the apt service
func NewAutotoolsService() IAutotools {
	return &SAutotools{}
}
