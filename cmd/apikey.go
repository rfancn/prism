package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdApikey = &cobra.Command{
	Use:   "apikey",
	Short: "API Key管理",
	Long:  `管理API密钥用于认证`,
}

var cmdApikeyList = &cobra.Command{
	Use:   "list",
	Short: "列出所有API Key",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("API Key列表功能将在后续实现")
	},
}

var cmdApikeyGenerate = &cobra.Command{
	Use:   "generate",
	Short: "生成新的API Key",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("生成API Key功能将在后续实现")
	},
}

var cmdApikeyDelete = &cobra.Command{
	Use:   "delete [id]",
	Short: "删除API Key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("请指定API Key ID")
			return
		}
		fmt.Printf("删除API Key功能将在后续实现: %s\n", args[0])
	},
}

func init() {
	cmdApikey.AddCommand(cmdApikeyList)
	cmdApikey.AddCommand(cmdApikeyGenerate)
	cmdApikey.AddCommand(cmdApikeyDelete)
	cmdApikey.Flags().StringVarP(&argConfigFile, "config", "c", "prism.toml", "配置文件路径")
}