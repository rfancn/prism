package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// Manager 插件管理器
// 负责加载、管理和调用插件
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*PluginInstance // 插件名称 -> 插件实例
	paths   []string                   // 插件搜索路径
}

// PluginInstance 插件实例
type PluginInstance struct {
	Name     string
	Path     string
	Client   *plugin.Client
	Instance RouterPlugin
	Info     *PluginInfo
}

// NewManager 创建新的插件管理器
func NewManager(pluginPaths []string) *Manager {
	return &Manager{
		plugins: make(map[string]*PluginInstance),
		paths:   pluginPaths,
	}
}

// LoadPlugin 加载单个插件
func (m *Manager) LoadPlugin(ctx context.Context, pluginPath string) (*PluginInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查插件文件是否存在
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("获取插件绝对路径失败: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("插件文件不存在: %s", absPath)
	}

	// 创建插件客户端配置
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          m.getPluginMap(),
		Cmd:              exec.Command(absPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           nil, // 可以传入自定义logger
	})

	// 连接插件进程
	grpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("连接插件失败: %w", err)
	}

	// 获取插件实例
	raw, err := grpcClient.Dispense(PluginName)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("获取插件实例失败: %w", err)
	}

	routerPlugin, ok := raw.(RouterPlugin)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("插件类型断言失败")
	}

	// 获取插件信息
	info, err := routerPlugin.Info(ctx)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("获取插件信息失败: %w", err)
	}

	instance := &PluginInstance{
		Name:     info.Name,
		Path:      absPath,
		Client:    client,
		Instance:  routerPlugin,
		Info:      info,
	}

	m.plugins[info.Name] = instance
	return instance, nil
}

// LoadAll 加载所有插件
func (m *Manager) LoadAll(ctx context.Context) error {
	for _, searchPath := range m.paths {
		// 遍历插件目录
		files, err := filepath.Glob(filepath.Join(searchPath, "*"))
		if err != nil {
			return fmt.Errorf("搜索插件目录失败: %w", err)
		}

		for _, file := range files {
			// 跳过目录
			if info, err := os.Stat(file); err != nil || info.IsDir() {
				continue
			}

			// 尝试加载插件
			_, err := m.LoadPlugin(ctx, file)
			if err != nil {
				// 记录错误但继续加载其他插件
				fmt.Printf("加载插件失败 %s: %v\n", file, err)
				continue
			}
		}
	}
	return nil
}

// GetPlugin 获取插件实例
func (m *Manager) GetPlugin(name string) (*PluginInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.plugins[name]
	return instance, ok
}

// GetAllPlugins 获取所有插件实例
func (m *Manager) GetAllPlugins() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*PluginInstance, 0, len(m.plugins))
	for _, instance := range m.plugins {
		plugins = append(plugins, instance)
	}
	return plugins
}

// UnloadPlugin 卸载插件
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("插件不存在: %s", name)
	}

	instance.Client.Kill()
	delete(m.plugins, name)
	return nil
}

// UnloadAll 卸载所有插件
func (m *Manager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, instance := range m.plugins {
		instance.Client.Kill()
		delete(m.plugins, name)
	}
}

// Match 使用所有插件尝试匹配请求
// 返回第一个匹配成功的插件结果
func (m *Manager) Match(ctx context.Context, req *MatchRequest) (*PluginInstance, *MatchResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, instance := range m.plugins {
		resp, err := instance.Instance.Match(ctx, req)
		if err != nil {
			// 记录错误但继续尝试其他插件
			fmt.Printf("插件 %s 匹配失败: %v\n", instance.Name, err)
			continue
		}

		if resp.Matched {
			return instance, resp, nil
		}
	}

	return nil, &MatchResponse{Matched: false}, nil
}

// MatchWithPlugin 使用指定插件匹配请求
func (m *Manager) MatchWithPlugin(ctx context.Context, pluginName string, req *MatchRequest) (*MatchResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", pluginName)
	}

	return instance.Instance.Match(ctx, req)
}

// GetPluginInfo 获取插件信息
func (m *Manager) GetPluginInfo(ctx context.Context, name string) (*PluginInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", name)
	}

	return instance.Instance.Info(ctx)
}

// ListPlugins 列出所有插件信息
func (m *Manager) ListPlugins() []*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]*PluginInfo, 0, len(m.plugins))
	for _, instance := range m.plugins {
		infos = append(infos, instance.Info)
	}
	return infos
}

// AddPluginPath 添加插件搜索路径
func (m *Manager) AddPluginPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.paths = append(m.paths, path)
}

// getPluginMap 返回插件映射
func (m *Manager) getPluginMap() map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		PluginName: &GRPCPlugin{},
	}
}

// ReloadPlugin 重新加载插件
func (m *Manager) ReloadPlugin(ctx context.Context, name string) (*PluginInstance, error) {
	m.mu.RLock()
	instance, ok := m.plugins[name]
	path := ""
	if ok {
		path = instance.Path
	}
	m.mu.RUnlock()

	if path == "" {
		return nil, fmt.Errorf("插件不存在: %s", name)
	}

	// 先卸载
	if err := m.UnloadPlugin(name); err != nil {
		return nil, err
	}

	// 重新加载
	return m.LoadPlugin(ctx, path)
}