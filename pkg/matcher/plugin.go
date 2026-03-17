package matcher

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/plugin"
)

// PluginMatcher 插件匹配器
// 用于调用外部插件执行匹配逻辑
type PluginMatcher struct {
	pluginMgr *plugin.Manager
}

// NewPluginMatcher 创建插件匹配器
func NewPluginMatcher(pluginMgr *plugin.Manager) *PluginMatcher {
	return &PluginMatcher{
		pluginMgr: pluginMgr,
	}
}

// Match 执行插件匹配
func (m *PluginMatcher) Match(c *gin.Context, rule *db.RouteRule) Result {
	// 1. 检查规则是否有效
	if rule == nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrNilRule,
		}
	}

	// 2. 检查插件管理器
	if m.pluginMgr == nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrNilPluginManager,
		}
	}

	// 3. 获取插件名称（必需）
	pluginName := rule.PluginName
	if !pluginName.Valid || pluginName.String == "" {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrEmptyPluginName,
		}
	}

	// 4. 构建匹配请求
	req := buildMatchRequest(c)

	// 5. 调用插件执行匹配
	ctx := context.Background()
	resp, err := m.pluginMgr.MatchWithPlugin(ctx, pluginName.String, req)
	if err != nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   err,
		}
	}

	// 6. 返回匹配结果
	return Result{
		Matched: resp.Matched,
		Params:  resp.Params,
		Error:   nil,
	}
}

// buildMatchRequest 构建插件匹配请求
func buildMatchRequest(c *gin.Context) *plugin.MatchRequest {
	// 获取请求体
	body := []byte{}
	if c.Request.Body != nil {
		bodyData, err := c.GetRawData()
		if err == nil {
			body = bodyData
		}
		// 恢复请求体
		c.Request.Body = c.Request.Body
	}

	// 构建请求
	req := &plugin.MatchRequest{
		Method:     c.Request.Method,
		Headers:    getHeaders(c),
		URLParams:  getURLParams(c),
		PathParams: getPathParamsFromGin(c),
		Body:       body,
		RemoteAddr: c.ClientIP(),
		Path:       c.Request.URL.Path,
	}

	return req
}

// getPathParamsFromGin 从Gin上下文获取路径参数
func getPathParamsFromGin(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for _, param := range c.Params {
		params[param.Key] = param.Value
	}
	return params
}