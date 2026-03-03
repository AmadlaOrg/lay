package container

import "os"

// For mocking
var osGetenv = os.Getenv

// Detector defines the interface for container runtime detection
type Detector interface {
	Detect(runtimeOverride string) (string, error)
}

// NewDetectorService creates a new detector service
func NewDetectorService() Detector {
	return &DetectorService{}
}
