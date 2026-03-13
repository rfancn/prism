// Package cmd contains CLI commands for Prism.
package cmd

import (
	"github.com/rfancn/prism/g"
	panicUtils "github.com/hdget/utils/panic"
	"github.com/hdget/utils/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Long:  "Prism - HTTP/HTTPS 请求中继工具",
	Short: "HTTP/HTTPS 请求中继工具",
	Use:   "prism",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&g.Debug, "debug", "", true, "debug mode")

	// Add subcommands
	rootCmd.AddCommand(cmdRun)
	rootCmd.AddCommand(cmdTui)
	rootCmd.AddCommand(cmdRoute)
	rootCmd.AddCommand(cmdApikey)
	rootCmd.AddCommand(cmdVersion)
}

// Execute runs the root command.
func Execute() {
	defer func() {
		if r := recover(); r != nil {
			panicUtils.RecordErrorStack(g.App)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("root command execute", "err", err)
	}
}