package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build information (set by ldflags during build)
var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Prism - HTTP/HTTPS 请求中继工具\n")
		fmt.Printf("版本:     %s\n", Version)
		fmt.Printf("提交:     %s\n", GitCommit)
		fmt.Printf("构建时间: %s\n", BuildDate)
	},
}