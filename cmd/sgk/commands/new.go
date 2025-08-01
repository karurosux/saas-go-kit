package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewProjectFunc is the function signature for creating projects
type NewProjectFunc func(projectName string, modules []string, goModule, database string) error

// NewCmd creates the new command
func NewCmd(createProject NewProjectFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new SaaS project with selected modules",
		Long:  "Creates a complete SaaS project with go.mod, main.go, and selected modules",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			
			// Get flags
			modules, _ := cmd.Flags().GetStringSlice("modules")
			goModule, _ := cmd.Flags().GetString("go-module")
			database, _ := cmd.Flags().GetString("database")
			
			if err := createProject(projectName, modules, goModule, database); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Project '%s' created successfully!\n", projectName)
			fmt.Printf("📂 cd %s && go mod tidy && go run main.go\n", projectName)
		},
	}

	// Add flags
	cmd.Flags().StringSlice("modules", []string{"auth"}, "Modules to include (auth,subscription,team,etc)")
	cmd.Flags().String("go-module", "", "Go module path (defaults to project name)")
	cmd.Flags().String("database", "postgres", "Database type (postgres, mysql, sqlite)")

	return cmd
}