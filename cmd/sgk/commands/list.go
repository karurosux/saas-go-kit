package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ListModulesFunc is the function signature for listing modules
type ListModulesFunc func() 

// ListInstalledModulesFunc is the function signature for listing installed modules
type ListInstalledModulesFunc func() error

// ListCmd creates the list command
func ListCmd(listModules ListModulesFunc, listInstalled ListInstalledModulesFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available or installed modules",
		Run: func(cmd *cobra.Command, args []string) {
			installed, _ := cmd.Flags().GetBool("installed")
			
			if installed {
				if err := listInstalled(); err != nil {
					fmt.Fprintf(os.Stderr, "Error listing modules: %v\n", err)
					os.Exit(1)
				}
			} else {
				listModules()
			}
		},
	}

	cmd.Flags().Bool("installed", false, "List only installed modules")
	return cmd
}