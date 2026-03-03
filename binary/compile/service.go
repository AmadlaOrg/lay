package compile

// NewDetectorService creates a new build system detector
func NewDetectorService() Detector {
	return &DetectorService{}
}
