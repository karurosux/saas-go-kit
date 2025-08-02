package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RemoveModuleFunc func(moduleName string) error

func RemoveCmd(removeModule RemoveModuleFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [module]",
		Short: "Remove a module from your project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			
			if err := removeModule(moduleName); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing module: %v\n", err)
				os.Exit(1)
			}
		},
	}
}