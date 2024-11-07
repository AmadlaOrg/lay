package linux

import "fmt"

type Values int

// Enum values
const (
	choco Values = iota
	scoop
	winget
)

// String method to convert enum values to their string representation
func (r Values) String() string {
	values := [...]string{
		"choco",
		"scoop",
		"winget",
	}

	if int(r) < 0 || int(r) >= len(values) {
		return fmt.Sprintf("Unknown(%d)", r) // Handle invalid enum values
	}
	return values[r]
}

// Exists checks if a given package manager name exist
func Exists(name string) bool {
	for _, v := range [...]Values{choco, scoop, winget} {
		if v.String() == name {
			return true
		}
	}
	return false
}
