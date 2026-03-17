// Example Plugin - 示例路由插件
//
// 这是一个示例插件，展示如何实现 RouterPlugin 接口。
// 插件通过路径前缀匹配请求，并提取租户标识符。
//
// 构建命令:
//   go build -o example-plugin .
//
// 使用方法:
//   将编译后的二进制文件放到插件目录中，Prism 会自动加载。
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/rfancn/prism/plugin"
)

// ExamplePlugin 示例插件实现
type ExamplePlugin struct{}

// Info 返回插件信息
func (p *ExamplePlugin) Info(ctx context.Context) (*plugin.PluginInfo, error) {
	return &plugin.PluginInfo{
		Name:        "example-plugin",
		Description: "示例路由插件，通过路径前缀匹配请求",
		Version:     "1.0.0",
		Author:      "Prism Team",
	}, nil
}

// Match 执行路由匹配
// 示例匹配规则：
// - 匹配路径以 /api/tenant/ 开头的请求
// - 从路径中提取租户ID
// - 支持特定请求头过滤
func (p *ExamplePlugin) Match(ctx context.Context, req *plugin.MatchRequest) (*plugin.MatchResponse, error) {
	// 检查路径是否以 /api/tenant/ 开头
	if !strings.HasPrefix(req.Path, "/api/tenant/") {
		return &plugin.MatchResponse{
			Matched: false,
		}, nil
	}

	// 检查请求方法（可选）
	if req.Method != "GET" && req.Method != "POST" {
		return &plugin.MatchResponse{
			Matched: false,
			Error:   "只支持 GET 和 POST 请求",
		}, nil
	}

	// 检查特定的请求头（可选）
	authHeader := req.Headers["Authorization"]
	if authHeader == "" {
		return &plugin.MatchResponse{
			Matched: false,
			Error:   "缺少 Authorization 请求头",
		}, nil
	}

	// 从路径中提取租户ID
	// 路径格式: /api/tenant/{tenant_id}/...
	pathParts := strings.Split(strings.TrimPrefix(req.Path, "/api/tenant/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		return &plugin.MatchResponse{
			Matched: false,
			Error:   "无法从路径中提取租户ID",
		}, nil
	}

	tenantID := pathParts[0]

	// 返回匹配成功
	return &plugin.MatchResponse{
		Matched: true,
		Params: map[string]string{
			"tenant_id": tenantID,
		},
		RouteID: fmt.Sprintf("tenant-%s", tenantID),
	}, nil
}

func main() {
	// 启动插件服务
	plugin.Serve(&ExamplePlugin{})
}