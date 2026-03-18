package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
)

// PluginName 插件名称常量
const PluginName = "router-plugin"

// Handshake 握手配置，用于验证客户端和插件之间的连接
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ROUTER_PLUGIN",
	MagicCookieValue: "PRISM_ROUTER_PLUGIN_1.0",
}

// MatchRequest 匹配请求结构体
type MatchRequest struct {
	Method     string            // HTTP方法
	Headers    map[string]string // HTTP请求头
	URLParams  map[string]string // URL查询参数
	PathParams map[string]string // 路径参数
	Body       []byte            // 请求体
	RemoteAddr string            // 客户端地址
	Path       string            // 请求路径
}

// MatchResponse 匹配响应结构体
type MatchResponse struct {
	Matched bool              // 是否匹配成功
	Params  map[string]string // 提取的参数
	Error   string            // 错误信息
	RouteID string            // 匹配的路由ID
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string // 插件名称
	Description string // 插件描述
	Version     string // 插件版本
	Author      string // 作者
}

// RouterPlugin 路由插件接口
// 插件开发者需要实现这个接口
type RouterPlugin interface {
	// Info 返回插件信息
	Info(ctx context.Context) (*PluginInfo, error)

	// Match 执行路由匹配
	// 返回 MatchResponse 指示是否匹配成功以及提取的参数
	Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error)
}