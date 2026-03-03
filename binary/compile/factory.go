package compile

import (
	"fmt"

	"github.com/AmadlaOrg/lay/binary/compile/autotools"
	"github.com/AmadlaOrg/lay/binary/compile/cargo"
	"github.com/AmadlaOrg/lay/binary/compile/cmake"
	"github.com/AmadlaOrg/lay/binary/compile/golang"
	"github.com/AmadlaOrg/lay/binary/compile/makefile"
	"github.com/AmadlaOrg/lay/binary/compile/meson"
)

// NewBuildSystemByName returns a build system implementation by name
func NewBuildSystemByName(name string) (BuildSystem, error) {
	switch name {
	case "autotools":
		return autotools.NewService(), nil
	case "cmake":
		return cmake.NewService(), nil
	case "meson":
		return meson.NewService(), nil
	case "makefile":
		return makefile.NewService(), nil
	case "cargo":
		return cargo.NewService(), nil
	case "golang":
		return golang.NewService(), nil
	default:
		return nil, fmt.Errorf("unsupported build system: %s", name)
	}
}
