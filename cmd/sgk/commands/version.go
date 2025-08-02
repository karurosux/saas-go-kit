package commands

import (
	"fmt"
	"runtime"
	"runtime/debug"

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
			version := Version
			commit := Commit
			buildDate := BuildDate
			
			if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
				if info.Main.Version != "" && info.Main.Version != "(devel)" {
					version = info.Main.Version
				}
				
				for _, setting := range info.Settings {
					switch setting.Key {
					case "vcs.revision":
						if commit == "none" && setting.Value != "" {
							commit = setting.Value
							if len(commit) > 7 {
								commit = commit[:7]
							}
						}
					case "vcs.time":
						if buildDate == "unknown" && setting.Value != "" {
							buildDate = setting.Value
						}
					case "vcs.modified":
						if setting.Value == "true" && version != "dev" {
							version += "-dirty"
						}
					}
				}
			}
			
			fmt.Printf("SGK (SaaS Go Kit) CLI\n")
			fmt.Printf("Version:    %s\n", version)
			fmt.Printf("Commit:     %s\n", commit)
			fmt.Printf("Built:      %s\n", buildDate)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}