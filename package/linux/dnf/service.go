package dnf

// NewDnfService to set up the dnf service
func NewDnfService() IDnf {
	return &SDnf{}
}
