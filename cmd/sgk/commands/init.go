package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type InitProjectFunc func() error

func InitCmd(initProject InitProjectFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project with sgk support",
		Long:  "Creates sgk.json configuration and sets up basic project structure",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initProject(); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Project initialized successfully!")
		},
	}
}