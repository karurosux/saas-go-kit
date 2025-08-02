package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type CrudGeneratorFunc func(moduleName string) error

func CrudCmd(generateCrud CrudGeneratorFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "crud [module-name]",
		Short: "Generate a complete CRUD module with model, repository, service, and controller",
		Long: `Generate a complete CRUD module

Example: sgk crud product`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			moduleName := strings.ToLower(strings.TrimSpace(args[0]))
			
			if moduleName == "" {
				fmt.Fprintf(os.Stderr, "Error: module name cannot be empty\n")
				os.Exit(1)
			}
			
			reservedNames := []string{"core", "main", "internal", "auth", "role", "health", "subscription"}
			for _, reserved := range reservedNames {
				if moduleName == reserved {
					fmt.Fprintf(os.Stderr, "Error: '%s' is a reserved module name\n", moduleName)
					os.Exit(1)
				}
			}
			
			if err := generateCrud(moduleName); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating CRUD module: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Printf("✅ CRUD module '%s' generated successfully!\n", moduleName)
			fmt.Printf("📁 Files created in internal/%s/\n", moduleName)
			fmt.Printf("🔄 Module registered in main.go\n")
		},
	}
}