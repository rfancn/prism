package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hdget/sdk"
	"github.com/hdget/sdk/providers/db/sqlite3/sqlc"
	"github.com/hdget/utils/logger"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/repository"
	"github.com/spf13/cobra"
)

var cmdRoute = &cobra.Command{
	Use:   "route",
	Short: "路由管理",
	Long:  `查看系统路由配置（路由由来源、项目和路由规则自动生成）`,
}

var cmdRouteList = &cobra.Command{
	Use:   "list",
	Short: "列出所有路由",
	Long: `列出系统中所有启用的路由规则。

路由是自动生成的，基于以下三层结构：
  来源(Source) -> 项目(Project) -> 路由规则(RouteRule)

要修改路由，请通过以下方式：
  - 使用 'prism tui' 在 TUI 界面中管理
  - 直接操作数据库中的 source、project、route_rule 表`,
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
		listRoutes()
	},
}

func init() {
	cmdRoute.AddCommand(cmdRouteList)
	cmdRouteList.Flags().StringVar(&argDbPath, "db", g.DefaultDbPath, "数据库文件路径")
}

// listRoutes 列出所有路由
func listRoutes() {
	// Run schema migrations to ensure tables exist
	if err := runMigrations(sdk.Db().My().SqlDB()); err != nil {
		sdk.Logger().Fatal("数据库迁移失败", "err", err)
	}

	queries := repository.New()
	ctx := context.Background()

	// 加载所有来源
	sources, err := queries.ListSources(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载来源失败: %v\n", err)
		os.Exit(1)
	}

	if len(sources) == 0 {
		fmt.Println("暂无路由配置")
		fmt.Println("\n提示: 使用 'prism tui' 创建来源、项目和路由规则")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "来源\t项目\t规则名称\t匹配类型\t路径模式\t目标URL\t优先级")
	fmt.Fprintln(w, "────\t────\t──────\t──────\t──────\t──────\t──────")

	totalRules := 0
	for _, source := range sources {
		// 加载该来源下的项目
		projects, err := queries.ListProjectsBySourceID(ctx, source.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载项目失败 (source=%s): %v\n", source.Name, err)
			continue
		}

		for _, project := range projects {
			// 加载该项目下的路由规则
			rules, err := queries.ListRouteRulesByProjectID(ctx, project.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "加载路由规则失败 (project=%s): %v\n", project.Name, err)
				continue
			}

			for _, rule := range rules {
				pathPattern := "-"
				if rule.PathPattern.Valid {
					pathPattern = rule.PathPattern.String
				}

				priority := "0"
				if rule.Priority.Valid {
					priority = fmt.Sprintf("%d", rule.Priority.Int64)
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					source.Name,
					project.Name,
					rule.Name,
					rule.MatchType,
					pathPattern,
					truncateURL(project.TargetUrl.String, 40),
					priority,
				)
				totalRules++
			}
		}
	}

	w.Flush()

	fmt.Printf("\n总计: %d 条路由规则\n", totalRules)
	if totalRules == 0 {
		fmt.Println("\n提示: 使用 'prism tui' 创建来源、项目和路由规则")
	}
}

// truncateURL 截断URL以便显示
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}