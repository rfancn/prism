package cmd

import (
	"context"
	"fmt"

	"github.com/hdget/sdk"
	"github.com/hdget/sdk/providers/db/sqlite3/sqlc"
	"github.com/hdget/utils/logger"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/pkg/config"
	"github.com/rfancn/prism/tui"
	"github.com/spf13/cobra"
)

var cmdTui = &cobra.Command{
	Use:   "tui",
	Short: "启动TUI管理界面",
	Long:  `启动终端用户界面(TUI)管理Prism配置`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 构建配置内容，使用命令行指定的数据库路径
		configContent := []byte(fmt.Sprintf(`
[sdk]
[sdk.sqlite]
db = "%s"
`, argDbPath))

		// Initialize SDK with config content (not config file)
		err := sdk.New(g.App, sdk.WithConfigContent(configContent)).
			Initialize(
				sqlc.Capability,
			)
		if err != nil {
			logger.Fatal("初始化SDK失败", "err", err)
		}

		// 从数据库加载应用配置到全局配置
		configManager := config.NewConfigManager()
		appConfig, err := configManager.LoadAppConfig(context.Background())
		if err != nil {
			logger.Fatal("加载应用配置失败", "err", err)
		}
		g.Config.App = *appConfig
	},
	Run: func(cmd *cobra.Command, args []string) {
		runTui()
	},
}

func init() {
	cmdTui.Flags().StringVar(&argDbPath, "db", g.DefaultDbPath, "数据库文件路径")
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