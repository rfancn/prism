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
	"github.com/rfancn/prism/pkg/proxy"
	"github.com/rfancn/prism/pkg/types"
	"github.com/rfancn/prism/plugin"
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
func NewRouter(pluginPaths []string) (*Router, error) {
	// 创建CEL引擎
	celEngine, err := cel.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL engine: %w", err)
	}

	// 创建插件管理器
	pluginMgr := plugin.NewManager(pluginPaths)

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
	for _, projectConfig := range sourceConfig.Projects {
		// 按优先级排序规则（已经在SQL中按priority排序，这里再确认）
		rules := projectConfig.Rules
		sort.Slice(rules, func(i, j int) bool {
			// priority小的优先
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
	// 构建转发选项
	opts := &proxy.ForwardOptions{
		TargetURL: matchCtx.Rule.TargetUrl,
		Params:    matchCtx.Params,
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