package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// GenerateClientsFunc is the function signature for generating clients
type GenerateClientsFunc func() error

// GenerateCmd creates the generate command
func GenerateCmd(generateClients GenerateClientsFunc) *cobra.Command {
	return &cobra.Command{
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
}