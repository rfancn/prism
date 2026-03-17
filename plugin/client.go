package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GRPCClient 实现RouterPlugin接口的gRPC客户端
type GRPCClient struct {
	client RouterPluginServiceClient
}

// NewGRPCClient 创建新的gRPC客户端
func NewGRPCClient(conn grpc.ClientConnInterface) *GRPCClient {
	return &GRPCClient{
		client: NewRouterPluginServiceClient(conn),
	}
}

// Info 获取插件信息
func (c *GRPCClient) Info(ctx context.Context) (*PluginInfo, error) {
	// 发送空请求
	resp, err := c.client.Info(ctx, &PluginInfo{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Match 执行路由匹配
func (c *GRPCClient) Match(ctx context.Context, req *MatchRequest) (*MatchResponse, error) {
	return c.client.Match(ctx, req)
}

// GRPCPlugin 实现hashicorp/go-plugin的GRPCPlugin接口
type GRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl RouterPlugin
}

// GRPCServer 返回gRPC服务端
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterRouterPluginServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient 返回gRPC客户端
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return NewGRPCClient(c), nil
}

// 确保GRPCClient实现了RouterPlugin接口
var _ RouterPlugin = (*GRPCClient)(nil)