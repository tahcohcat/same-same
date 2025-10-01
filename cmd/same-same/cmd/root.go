package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose   bool
	namespace string
	dryRun    bool
)

var rootCmd = &cobra.Command{
	Use:   "same-same",
	Short: "Same-Same Vector Database Microservice",
	Long: `Same-Same is a lightweight RESTful microservice for storing and searching 
vectors using cosine similarity, with built-in embedding generation for text.

Designed and optimized for quick prototyping and exploration of the vector 
space with minimal setup requirements.`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace for vectors")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "perform a dry run without making changes")
}
