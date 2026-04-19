package container

import "os"

// For mocking
var osGetenv = os.Getenv

// Detector defines the interface for container runtime detection
type Detector interface {
	Detect(runtimeOverride string) (string, error)
}

// New creates a new Detector.
func New() Detector {
	return &DetectorService{}
}
