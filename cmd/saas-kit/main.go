package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "saas-kit",
		Short: "SaaS Go Kit - Copy-paste modular components for Go SaaS applications",
		Long: `SaaS Go Kit CLI allows you to add pre-built, customizable modules to your Go application.
Similar to shadcn/ui, you own the code and can modify it as needed.

Available modules: auth, subscription, team, notification, health, role, job, sse, container`,
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(generateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project with saas-kit support",
	Long:  "Creates saas-kit.json configuration and sets up basic project structure",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initProject(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Project initialized successfully!")
	},
}

var addCmd = &cobra.Command{
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available or installed modules",
	Run: func(cmd *cobra.Command, args []string) {
		installed, _ := cmd.Flags().GetBool("installed")
		
		if installed {
			if err := listInstalledModules(); err != nil {
				fmt.Fprintf(os.Stderr, "Error listing modules: %v\n", err)
				os.Exit(1)
			}
		} else {
			listAvailableModules()
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [module]",
	Short: "Remove a module from your project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		moduleName := args[0]
		
		if err := removeModule(moduleName); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing module: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Module '%s' removed successfully!\n", moduleName)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [module]",
	Short: "Update a module to the latest version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		moduleName := args[0]
		
		if err := updateModule(moduleName); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating module: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Module '%s' updated successfully!\n", moduleName)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate TypeScript clients and other assets",
	Run: func(cmd *cobra.Command, args []string) {
		if err := generateClients(); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating clients: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ TypeScript clients generated successfully!")
	},
}

func init() {
	// Add command flags
	addCmd.Flags().String("database", "postgres", "Database type (postgres, mysql, sqlite)")
	addCmd.Flags().String("route-prefix", "", "Route prefix for the module")
	addCmd.Flags().String("email-provider", "smtp", "Email provider (smtp, sendgrid, ses)")
	addCmd.Flags().String("payment-provider", "stripe", "Payment provider (stripe)")
	addCmd.Flags().Bool("require-verification", false, "Require email verification")
	
	listCmd.Flags().Bool("installed", false, "List only installed modules")
}