package cmd

import (
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/tui"

	"github.com/hdget/sdk"
	"github.com/hdget/sdk/providers/db/sqlite3/sqlc"
	"github.com/hdget/utils/logger"
	"github.com/spf13/cobra"
)

var cmdTui = &cobra.Command{
	Use:   "tui",
	Short: "启动TUI管理界面",
	Long:  `启动终端用户界面(TUI)管理Prism配置`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Initialize SDK with config file
		err := sdk.New(g.App, sdk.WithConfigFile(argConfigFile)).
			UseConfig(&g.Config).
			Initialize(
				sqlc.Capability,
			)
		if err != nil {
			logger.Fatal("初始化SDK失败", "err", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		runTui()
	},
}

func init() {
	cmdTui.Flags().StringVarP(&argConfigFile, "config", "c", "prism.toml", "配置文件路径")
}

func runTui() {
	// Run schema migrations
	if err := runMigrations(sdk.Db().My().SqlDB()); err != nil {
		sdk.Logger().Fatal("数据库迁移失败", "err", err)
	}

	// Run TUI
	if err := tui.Run(); err != nil {
		sdk.Logger().Error("TUI错误", "err", err)
	}
}