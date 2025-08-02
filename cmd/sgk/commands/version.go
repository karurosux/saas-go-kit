package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Display version, commit hash, build date, and runtime information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SGK (SaaS Go Kit) CLI\n")
			fmt.Printf("Version:    %s\n", Version)
			fmt.Printf("Commit:     %s\n", Commit)
			fmt.Printf("Built:      %s\n", BuildDate)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}