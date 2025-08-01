package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// UpdateModuleFunc is the function signature for updating modules
type UpdateModuleFunc func(moduleName string) error

// UpdateCmd creates the update command
func UpdateCmd(updateModule UpdateModuleFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "update [module]",
		Short: "Update a module to the latest version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			
			if err := updateModule(moduleName); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating module: %v\n", err)
				os.Exit(1)
			}
		},
	}
}