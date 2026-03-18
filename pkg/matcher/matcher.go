// Package matcher 提供请求匹配器实现
// 支持多种匹配类型：参数化路径、URL参数、请求体、表单数据、插件
package matcher

import (
	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
	"github.com/rfancn/prism/pkg/plugin"
)

// Result 匹配结果
type Result struct {
	// Matched 是否匹配成功
	Matched bool
	// Params 提取的参数
	Params map[string]string
	// Error 匹配过程中的错误
	Error error
}

// Matcher 匹配器接口
type Matcher interface {
	// Match 执行匹配
	// c: Gin上下文
	// rule: 路由规则
	// 返回匹配结果
	Match(c *gin.Context, rule *db.RouteRule) Result
}

// Factory 创建匹配器的工厂
type Factory struct {
	celEngine *cel.Engine
	pluginMgr *plugin.Manager
}

// NewFactory 创建匹配器工厂
func NewFactory(celEngine *cel.Engine, pluginMgr *plugin.Manager) *Factory {
	return &Factory{
		celEngine: celEngine,
		pluginMgr: pluginMgr,
	}
}

// Create 根据匹配类型创建匹配器
func (f *Factory) Create(matchType string) Matcher {
	switch matchType {
	case MatchTypePathParam:
		return &ParamPathMatcher{celEngine: f.celEngine}
	case MatchTypeURLParam:
		return &URLParamMatcher{celEngine: f.celEngine}
	case MatchTypeRequestBody:
		return &RequestBodyMatcher{celEngine: f.celEngine}
	case MatchTypeRequestForm:
		return &RequestFormMatcher{celEngine: f.celEngine}
	case MatchTypePlugin:
		return &PluginMatcher{pluginMgr: f.pluginMgr}
	default:
		return nil
	}
}

// CreateWithEngine 创建带CEL引擎的匹配器（用于测试）
func (f *Factory) CreateWithEngine(matchType string, celEngine *cel.Engine) Matcher {
	switch matchType {
	case MatchTypePathParam:
		return &ParamPathMatcher{celEngine: celEngine}
	case MatchTypeURLParam:
		return &URLParamMatcher{celEngine: celEngine}
	case MatchTypeRequestBody:
		return &RequestBodyMatcher{celEngine: celEngine}
	case MatchTypeRequestForm:
		return &RequestFormMatcher{celEngine: celEngine}
	case MatchTypePlugin:
		return &PluginMatcher{pluginMgr: f.pluginMgr}
	default:
		return nil
	}
}

// 匹配类型常量
const (
	// MatchTypePathParam 参数化路径匹配
	MatchTypePathParam = "param_path"
	// MatchTypeURLParam URL参数匹配
	MatchTypeURLParam = "url_param"
	// MatchTypeRequestBody 请求体匹配（JSON）
	MatchTypeRequestBody = "request_body"
	// MatchTypeRequestForm 表单数据匹配
	MatchTypeRequestForm = "request_form"
	// MatchTypePlugin 插件匹配
	MatchTypePlugin = "plugin"
)

// ValidMatchTypes 返回所有有效的匹配类型
func ValidMatchTypes() []string {
	return []string{
		MatchTypePathParam,
		MatchTypeURLParam,
		MatchTypeRequestBody,
		MatchTypeRequestForm,
		MatchTypePlugin,
	}
}

// IsValidMatchType 检查匹配类型是否有效
func IsValidMatchType(matchType string) bool {
	for _, t := range ValidMatchTypes() {
		if t == matchType {
			return true
		}
	}
	return false
}
