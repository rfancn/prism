package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdRoute = &cobra.Command{
	Use:   "route",
	Short: "路由管理",
	Long:  `管理代理路由配置`,
}

var cmdRouteList = &cobra.Command{
	Use:   "list",
	Short: "列出所有路由",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("路由列表功能将在后续实现")
	},
}

var cmdRouteAdd = &cobra.Command{
	Use:   "add",
	Short: "添加路由",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("添加路由功能将在后续实现")
	},
}

var cmdRouteDelete = &cobra.Command{
	Use:   "delete [id]",
	Short: "删除路由",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("请指定路由ID")
			return
		}
		fmt.Printf("删除路由功能将在后续实现: %s\n", args[0])
	},
}

func init() {
	cmdRoute.AddCommand(cmdRouteList)
	cmdRoute.AddCommand(cmdRouteAdd)
	cmdRoute.AddCommand(cmdRouteDelete)
	cmdRoute.Flags().StringVarP(&argConfigFile, "config", "c", "prism.toml", "配置文件路径")
}