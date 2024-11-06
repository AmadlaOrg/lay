package rpm

// NewRpmService to set up the dpkg service
func NewRpmService() IRpm {
	return &SRpm{}
}
