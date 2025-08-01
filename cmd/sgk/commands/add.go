package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// AddModuleFunc is the function signature for adding modules
type AddModuleFunc func(moduleName string, options map[string]interface{}) error

// AddCmd creates the add command
func AddCmd(addModule AddModuleFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [module]",
		Short: "Add a module to your project",
		Long:  "Copy a module template into your project with customization options",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := args[0]
			
			// Get flags
			database, _ := cmd.Flags().GetString("database")
			routePrefix, _ := cmd.Flags().GetString("route-prefix")
			
			options := map[string]interface{}{
				"database":     database,
				"route_prefix": routePrefix,
			}
			
			if err := addModule(moduleName, options); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding module: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Module '%s' added successfully!\n", moduleName)
		},
	}

	// Add flags
	cmd.Flags().String("database", "postgres", "Database type (postgres, mysql, sqlite)")
	cmd.Flags().String("route-prefix", "", "Route prefix for the module")

	return cmd
}