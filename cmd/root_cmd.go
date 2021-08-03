package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use: "supabase-admin-api",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

// RootCommand will setup and return the root command
func RootCommand() *cobra.Command {
	rootCmd.AddCommand(&serveCmd, &versionCmd)

	return &rootCmd
}
