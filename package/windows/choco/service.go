package choco

// NewChocoService to set up the dpkg service
func NewChocoService() IChoco {
	return &SChoco{}
}
