package router

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/pkg/cel"
	"github.com/rfancn/prism/pkg/matcher"
	"github.com/rfancn/prism/pkg/plugin"
	"github.com/rfancn/prism/pkg/proxy"
	"github.com/rfancn/prism/pkg/types"
)

// Router 路由管理器
// 负责路由匹配和请求转发
type Router struct {
	loader         *Loader
	matcherFactory *matcher.Factory
	pluginMgr      *plugin.Manager
	celEngine      *cel.Engine
	proxyHandler   *proxy.ProxyHandler
	config         *RouterConfig
}

// NewRouter 创建路由管理器
func NewRouter(pluginPath string) (*Router, error) {
	// 创建CEL引擎
	celEngine, err := cel.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL engine: %w", err)
	}

	// 创建插件管理器
	pluginMgr := plugin.NewManager(pluginPath)

	// 创建路由加载器
	loader := NewLoader()

	return &Router{
		loader:         loader,
		matcherFactory: matcher.NewFactory(celEngine, pluginMgr),
		pluginMgr:      pluginMgr,
		celEngine:      celEngine,
		proxyHandler:   proxy.NewProxyHandler(&types.TargetTLSConfig{}),
	}, nil
}

// LoadConfig 加载路由配置
func (r *Router) LoadConfig(ctx context.Context) error {
	config, err := r.loader.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load router config: %w", err)
	}
	r.config = config
	return nil
}

// LoadPlugins 加载插件
func (r *Router) LoadPlugins(ctx context.Context) error {
	return r.pluginMgr.LoadAll(ctx)
}

// Handler 返回Gin处理函数
func (r *Router) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. 提取来源名称（路径第一部分）
		sourceName := extractSourceName(c.Request.URL.Path)
		if sourceName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source"})
			return
		}

		// 2. 查找来源
		sourceConfig := r.findSource(sourceName)
		if sourceConfig == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}

		// 3. 遍历项目的路由规则，寻找匹配
		matchCtx := r.findMatch(ctx, c, sourceConfig)
		if matchCtx == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no matching rule"})
			return
		}

		// 4. 转发请求
		r.forwardRequest(c, matchCtx)
	}
}

// findSource 根据名称查找来源配置
func (r *Router) findSource(name string) *SourceConfig {
	if r.config == nil {
		return nil
	}

	for _, sourceConfig := range r.config.Sources {
		if sourceConfig.Source.Name == name {
			return sourceConfig
		}
	}
	return nil
}

// findMatch 在来源配置中查找匹配的路由规则
func (r *Router) findMatch(ctx context.Context, c *gin.Context, sourceConfig *SourceConfig) *MatchContext {
	// 复制项目列表并按优先级排序
	projects := make([]*ProjectConfig, len(sourceConfig.Projects))
	copy(projects, sourceConfig.Projects)
	sort.Slice(projects, func(i, j int) bool {
		// priority小的优先
		pi := projects[i].Project.Priority.Int64
		pj := projects[j].Project.Priority.Int64
		return pi < pj
	})

	for _, projectConfig := range projects {
		// 按优先级排序规则
		rules := projectConfig.Rules
		sort.Slice(rules, func(i, j int) bool {
			pi := rules[i].Rule.Priority.Int64
			pj := rules[j].Rule.Priority.Int64
			return pi < pj
		})

		for _, ruleConfig := range rules {
			rule := ruleConfig.Rule

			// 创建匹配器
			m := r.matcherFactory.Create(rule.MatchType)
			if m == nil {
				sdk.Logger().Debug("unsupported match type", "rule_id", rule.ID, "match_type", rule.MatchType)
				continue
			}

			// 执行匹配
			result := m.Match(c, rule)
			if result.Error != nil {
				sdk.Logger().Error("matcher error", "rule_id", rule.ID, "err", result.Error)
				continue
			}

			if result.Matched {
				sdk.Logger().Debug("route matched",
					"source", sourceConfig.Source.Name,
					"project", projectConfig.Project.Name,
					"rule", rule.Name,
					"params", result.Params,
				)

				return &MatchContext{
					Source:  sourceConfig.Source,
					Project: projectConfig.Project,
					Rule:    rule,
					Params:  result.Params,
				}
			}
		}
	}

	return nil
}

// forwardRequest 转发请求到目标服务器
func (r *Router) forwardRequest(c *gin.Context, matchCtx *MatchContext) {
	// 从 Project 获取目标URL
	if !matchCtx.Project.TargetUrl.Valid || matchCtx.Project.TargetUrl.String == "" {
		sdk.Logger().Error("project has no target URL configured",
			"project_id", matchCtx.Project.ID,
			"project_name", matchCtx.Project.Name,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "project target URL not configured"})
		return
	}

	// 构建转发选项
	opts := &proxy.ForwardOptions{
		TargetURL:  matchCtx.Project.TargetUrl.String,
		Params:     matchCtx.Params,
		SourceName: matchCtx.Source.Name,
		ExtraHeaders: map[string]string{
			"X-Prism-Source":  matchCtx.Source.Name,
			"X-Prism-Project": matchCtx.Project.Name,
			"X-Prism-Rule":    matchCtx.Rule.Name,
		},
	}

	// 调用 proxy 包的 Forward 方法
	if err := r.proxyHandler.Forward(c, opts); err != nil {
		sdk.Logger().Error("forward request failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// extractSourceName 从请求路径中提取来源名称
// 路径格式：/source_name/...
func extractSourceName(path string) string {
	// 去除前导斜杠
	path = strings.TrimPrefix(path, "/")

	// 获取第一部分作为来源名称
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}

	return parts[0]
}

// GetPluginManager 获取插件管理器
func (r *Router) GetPluginManager() *plugin.Manager {
	return r.pluginMgr
}

// GetConfig 获取当前路由配置
func (r *Router) GetConfig() *RouterConfig {
	return r.config
}

// ReloadConfig 重新加载路由配置
func (r *Router) ReloadConfig(ctx context.Context) error {
	return r.LoadConfig(ctx)
}

// GetLoader 获取路由加载器
func (r *Router) GetLoader() *Loader {
	return r.loader
}

// RegisterRoutes 注册所有 param_path 类型的路由到 Gin 引擎
// 遍历配置中的所有 Source -> Project -> Rule，为 match_type 为 param_path 的规则注册路由
func (r *Router) RegisterRoutes(engine *gin.Engine) error {
	// 检查配置是否已加载
	if r.config == nil {
		return fmt.Errorf("router config not loaded, please call LoadConfig first")
	}

	// 记录注册的路由数量
	registeredCount := 0

	// 遍历所有来源
	for _, sourceConfig := range r.config.Sources {
		sourceName := sourceConfig.Source.Name

		// 遍历所有项目
		for _, projectConfig := range sourceConfig.Projects {
			// 遍历所有规则
			for _, ruleConfig := range projectConfig.Rules {
				rule := ruleConfig.Rule

				// 只处理 param_path 类型的规则
				if rule.MatchType != matcher.MatchTypePathParam {
					continue
				}

				// 检查路径模式是否存在
				if !rule.PathPattern.Valid || rule.PathPattern.String == "" {
					sdk.Logger().Warn("rule has empty path pattern, skipping",
						"rule_id", rule.ID,
						"rule_name", rule.Name,
					)
					continue
				}

				// 构建完整路径：/{source_name}{path_pattern}
				pathPattern := rule.PathPattern.String
				fullPath := "/" + sourceName + convertToGinPath(pathPattern)

				// 使用 r.Handler() 作为处理函数
				handler := r.Handler()

				// 注册 GET 和 POST 方法的路由
				engine.GET(fullPath, handler)
				engine.POST(fullPath, handler)

				registeredCount++

				sdk.Logger().Info("route registered",
					"source", sourceName,
					"project", projectConfig.Project.Name,
					"rule", rule.Name,
					"path", fullPath,
					"methods", "GET,POST",
				)
			}
		}
	}

	sdk.Logger().Info("routes registration completed", "total_registered", registeredCount)
	return nil
}

// convertToGinPath 将路径参数格式从 {param} 转换为 Gin 格式 :param
// 例如: /orders/{id} -> /orders/:id
//
//	/users/{userId}/orders/{orderId} -> /users/:userId/orders/:orderId
func convertToGinPath(pattern string) string {
	var result strings.Builder
	i := 0
	n := len(pattern)

	for i < n {
		if pattern[i] == '{' {
			// 找到花括号的结束位置
			j := i + 1
			for j < n && pattern[j] != '}' {
				j++
			}
			if j < n && pattern[j] == '}' {
				// 提取参数名（不包含花括号）
				paramName := pattern[i+1 : j]
				result.WriteString(":")
				result.WriteString(paramName)
				i = j + 1
				continue
			}
		}
		result.WriteByte(pattern[i])
		i++
	}

	return result.String()
}
