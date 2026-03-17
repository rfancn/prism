package cmd

import (
	"fmt"

	"github.com/hdget/sdk"
	"github.com/hdget/sdk/providers/db/sqlite3/sqlc"
	"github.com/hdget/utils/logger"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/pkg/server"
	"github.com/rfancn/prism/pkg/types"
	"github.com/spf13/cobra"
)

var (
	argConfigFile string
	argCertFile   string
	argKeyFile    string
	argHost       string
	argPort       int

	cmdRun = &cobra.Command{
		Use:   "run",
		Short: "启动代理服务",
		Long:  `启动 HTTP/HTTPS 请求中继代理服务`,
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
			runServer()
		},
	}
)

func init() {
	cmdRun.Flags().StringVarP(&argConfigFile, "config", "c", "prism.toml", "配置文件路径")
	cmdRun.Flags().StringVarP(&argHost, "host", "h", "127.0.0.1", "服务监听主机")
	cmdRun.Flags().IntVarP(&argPort, "port", "p", 8080, "服务监听端口")
	cmdRun.Flags().StringVar(&argCertFile, "cert", "", "SSL 证书文件路径")
	cmdRun.Flags().StringVar(&argKeyFile, "key", "", "SSL 私钥文件路径")
}

func runServer() {
	// Run schema migrations
	if err := runMigrations(sdk.Db().My().SqlDB()); err != nil {
		sdk.Logger().Fatal("数据库迁移失败", "err", err)
	}

	// Determine address
	address := fmt.Sprintf("%s:%d", argHost, argPort)

	// Build TLS config if certificate files are provided
	var tlsConfig *types.ServerTLSConfig
	if argCertFile != "" && argKeyFile != "" {
		tlsConfig = &types.ServerTLSConfig{
			Enabled:  true,
			CertFile: argCertFile,
			KeyFile:  argKeyFile,
		}
	}

	sdk.Logger().Info("启动Prism代理服务...",
		"address", address,
	)

	// Create and run server
	srv, err := server.New(address, tlsConfig)
	if err != nil {
		sdk.Logger().Fatal("创建服务器失败", "err", err)
	}

	srv.Run()
}
