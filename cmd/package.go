package cmd

import (
	"github.com/AmadlaOrg/lay/package/linux/apt"
	"github.com/AmadlaOrg/lay/package/linux/dnf"
	"github.com/AmadlaOrg/lay/package/linux/dpkg"
	"github.com/AmadlaOrg/lay/package/linux/rpm"
	"github.com/AmadlaOrg/lay/package/linux/yum"
	"github.com/AmadlaOrg/lay/package/windows/choco"
	"github.com/AmadlaOrg/lay/package/windows/scoop"
	"github.com/AmadlaOrg/lay/package/windows/winget"
	"github.com/spf13/cobra"
)

var PackageCmd = &cobra.Command{
	Use:   "package",
	Short: "Compile a binary to a Go source file",
	//Long:  `A longer description that spans multiple lines and likely`,
	//Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	// Linux
	PackageCmd.AddCommand(apt.PackageAptCmd)
	PackageCmd.AddCommand(dnf.PackageDnfCmd)
	PackageCmd.AddCommand(dpkg.PackageDpkgCmd)
	PackageCmd.AddCommand(rpm.PackageRpmCmd)
	PackageCmd.AddCommand(yum.PackageYumCmd)

	// For Windows
	PackageCmd.AddCommand(choco.PackageChocoCmd)
	PackageCmd.AddCommand(scoop.PackageScoopCmd)
	PackageCmd.AddCommand(winget.PackageWingetCmd)
}
