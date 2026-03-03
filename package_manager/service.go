package package_manager

import "os"

// For mocking
var osGetenv = os.Getenv

// Detector defines the interface for package manager detection
type Detector interface {
	Detect(managerOverride string) (string, error)
}

// NewDetectorService creates a new detector service
func NewDetectorService() Detector {
	return &DetectorService{}
}
