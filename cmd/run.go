package cmd

import (
	"context"
	"fmt"

	"github.com/hdget/sdk"
	"github.com/hdget/sdk/providers/db/sqlite3/sqlc"
	"github.com/hdget/utils/logger"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/pkg/config"
	"github.com/rfancn/prism/pkg/server"
	"github.com/rfancn/prism/pkg/types"
	"github.com/spf13/cobra"
)

var (
	argDbPath   string
	argCertFile string
	argKeyFile  string
	argHost     string
	argPort     int

	cmdRun = &cobra.Command{
		Use:   "run",
		Short: "启动代理服务",
		Long:  `启动 HTTP/HTTPS 请求中继代理服务`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// 生成包含数据库路径的配置内容
			configContent := generateSDKConfig(argDbPath)

			// Initialize SDK with dynamic config content
			err := sdk.New(g.App, sdk.WithConfigContent(configContent)).
				Initialize(
					sqlc.Capability,
				)
			if err != nil {
				logger.Fatal("初始化SDK失败", "err", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}
)

// generateSDKConfig 生成 SDK 配置内容（TOML 格式）
func generateSDKConfig(dbPath string) []byte {
	return []byte(fmt.Sprintf(`
[sdk]
[sdk.logger]
level = "info"
filename = "%s.log"
[sdk.logger.rotate]
max_age = 7
rotation_time = 24

[sdk.sqlite]
db = "%s"
`, g.App, dbPath))
}

func init() {
	cmdRun.Flags().StringVar(&argDbPath, "db", g.DefaultDbPath, "数据库文件路径")
	cmdRun.Flags().StringVarP(&argHost, "host", "h", "", "服务监听主机（覆盖数据库配置）")
	cmdRun.Flags().IntVarP(&argPort, "port", "p", 0, "服务监听端口（覆盖数据库配置）")
	cmdRun.Flags().StringVar(&argCertFile, "cert", "", "SSL 证书文件路径")
	cmdRun.Flags().StringVar(&argKeyFile, "key", "", "SSL 私钥文件路径")
}

func runServer() {
	ctx := context.Background()

	// Run schema migrations
	if err := runMigrations(sdk.Db().My().SqlDB()); err != nil {
		sdk.Logger().Fatal("数据库迁移失败", "err", err)
	}

	// 从数据库加载应用配置
	configManager := config.NewConfigManager()
	appConfig, err := configManager.LoadAppConfig(ctx)
	if err != nil {
		sdk.Logger().Fatal("加载应用配置失败", "err", err)
	}

	// Build TLS config if certificate files are provided
	var tlsConfig *types.ServerTLSConfig
	if argCertFile != "" && argKeyFile != "" {
		tlsConfig = &types.ServerTLSConfig{
			Enabled:  true,
			CertFile: argCertFile,
			KeyFile:  argKeyFile,
		}
	}

	// 命令行参数覆盖数据库配置（优先级最高）
	if argHost != "" {
		appConfig.Server.Host = argHost
	}
	if argPort != 0 {
		appConfig.Server.Port = argPort
	}

	host := appConfig.Server.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := appConfig.Server.Port
	if port == 0 {
		port = 8080
	}

	sdk.Logger().Info("启动Prism代理服务...",
		"address", fmt.Sprintf("%s:%d", host, port),
		"db", argDbPath,
	)

	// Create and run server
	srv, err := server.New(appConfig, tlsConfig)
	if err != nil {
		sdk.Logger().Fatal("创建服务器失败", "err", err)
	}

	srv.Run()
}
