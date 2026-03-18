package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
)

// GRPCServer 实现gRPC服务端
type GRPCServer struct {
	UnimplementedRouterPluginServiceServer
	Impl RouterPlugin
}

// NewGRPCServer 创建新的gRPC服务端
func NewGRPCServer(impl RouterPlugin) *GRPCServer {
	return &GRPCServer{Impl: impl}
}

// Info 返回插件信息
func (s *GRPCServer) Info(ctx context.Context, req *PluginInfo) (*PluginInfo, error) {
	return s.Impl.Info(ctx)
}

// Match 执行路由匹配
func (s *GRPCServer) Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error) {
	return s.Impl.Match(ctx, req)
}

// Serve 启动插件服务
// 这是插件进程的主入口，插件开发者应该在main函数中调用此函数
func Serve(impl RouterPlugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			PluginName: &GRPCPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// 确保GRPCServer实现了RouterPluginServiceServer接口
var _ RouterPluginServiceServer = (*GRPCServer)(nil)