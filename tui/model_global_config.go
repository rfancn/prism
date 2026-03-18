package tui

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/config"
	"github.com/rfancn/prism/repository"
)

// 全局配置项常量
const (
	ConfigKeyIPWhitelistEnabled = "ip_whitelist_enabled"
)

// GlobalConfigModel 全局配置模型
type GlobalConfigModel struct {
	list         list.Model
	items        []ConfigItem
	state        AppState
	width        int
	height       int
	keys         KeyMap
	selectedItem int

	// 配置数据 - 功能开关
	ipWhitelistEnabled bool
	tlsEnabled         bool
	autoCertEnabled    bool
	tlsDomains         string

	// 配置数据 - 应用配置（来自 app_config 表）
	serverHost    string
	serverPort    int
	serverTLSPort int
	proxyTimeouts proxyTimeouts
}

// proxyTimeouts 代理超时配置
type proxyTimeouts struct {
	readTimeout  int
	writeTimeout int
	idleTimeout  int
}

// ConfigItem 配置项
type ConfigItem struct {
	key         string
	title       string
	description string
	value       string
	valueType   string // "toggle", "text"
	enabled     bool
}

// FilterValue implements list.Item interface
func (c ConfigItem) FilterValue() string {
	return c.title
}

// Title implements list.DefaultItem interface
func (c ConfigItem) Title() string {
	return c.title
}

// Description implements list.DefaultItem interface
func (c ConfigItem) Description() string {
	return c.description
}

// MsgGlobalConfigLoaded 全局配置加载完成消息
type MsgGlobalConfigLoaded struct {
	// 功能开关
	IPWhitelistEnabled bool
	TLSEnabled         bool
	AutoCertEnabled    bool
	TLSDomains         string
	// 应用配置
	ServerHost    string
	ServerPort    int
	ServerTLSPort int
	ProxyTimeouts proxyTimeouts
}

// NewGlobalConfigModel 创建新的全局配置模型
func NewGlobalConfigModel() *GlobalConfigModel {
	m := &GlobalConfigModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "系统配置", 80, 20)
	return m
}

// Init 初始化模型
func (m *GlobalConfigModel) Init() tea.Cmd {
	return m.loadConfig()
}

// SetSize 设置尺寸
func (m *GlobalConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// loadConfig 加载配置
func (m *GlobalConfigModel) loadConfig() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}

		ctx := context.Background()

		// 加载 IP 白名单开关
		ipWhitelistEnabled := false
		ipConfig, err := queries.GetGlobalConfig(ctx, ConfigKeyIPWhitelistEnabled)
		if err == nil && ipConfig.Value == "true" {
			ipWhitelistEnabled = true
		}

		// 加载 TLS 配置
		tlsConfig, err := queries.GetTLSConfig(ctx)
		tlsEnabled := false
		autoCertEnabled := false
		tlsDomains := ""
		if err == nil {
			tlsEnabled = tlsConfig.Enabled.Int64 == 1
			autoCertEnabled = tlsConfig.AutoCert.Int64 == 1
			tlsDomains = tlsConfig.Domains.String
		}

		// 从 app_config 表加载应用配置
		configMgr := config.NewConfigManager()
		appConfig, _ := configMgr.LoadAppConfig(ctx)

		return MsgGlobalConfigLoaded{
			IPWhitelistEnabled: ipWhitelistEnabled,
			TLSEnabled:         tlsEnabled,
			AutoCertEnabled:    autoCertEnabled,
			TLSDomains:         tlsDomains,
			ServerHost:         appConfig.Server.Host,
			ServerPort:         appConfig.Server.Port,
			ServerTLSPort:      appConfig.Server.TLSPort,
			ProxyTimeouts: proxyTimeouts{
				readTimeout:  appConfig.Proxy.ReadTimeout,
				writeTimeout: appConfig.Proxy.WriteTimeout,
				idleTimeout:  appConfig.Proxy.IdleTimeout,
			},
		}
	}
}

// refreshList 刷新列表
func (m *GlobalConfigModel) refreshList() {
	m.items = []ConfigItem{
		// 服务器配置
		{
			key:         "server.host",
			title:       "服务主机",
			description: "服务监听的主机地址",
			value:       m.serverHost,
			valueType:   "text",
			enabled:     false,
		},
		{
			key:         "server.port",
			title:       "HTTP 端口",
			description: "HTTP 服务监听端口",
			value:       strconv.Itoa(m.serverPort),
			valueType:   "text",
			enabled:     false,
		},
		{
			key:         "server.tls_port",
			title:       "HTTPS 端口",
			description: "HTTPS 服务监听端口",
			value:       strconv.Itoa(m.serverTLSPort),
			valueType:   "text",
			enabled:     false,
		},
		// 代理配置
		{
			key:         "proxy.read_timeout",
			title:       "读超时",
			description: "代理读取超时时间（秒）",
			value:       strconv.Itoa(m.proxyTimeouts.readTimeout),
			valueType:   "text",
			enabled:     false,
		},
		{
			key:         "proxy.write_timeout",
			title:       "写超时",
			description: "代理写入超时时间（秒）",
			value:       strconv.Itoa(m.proxyTimeouts.writeTimeout),
			valueType:   "text",
			enabled:     false,
		},
		{
			key:         "proxy.idle_timeout",
			title:       "空闲超时",
			description: "代理空闲超时时间（秒）",
			value:       strconv.Itoa(m.proxyTimeouts.idleTimeout),
			valueType:   "text",
			enabled:     false,
		},
		// 功能开关
		{
			key:         ConfigKeyIPWhitelistEnabled,
			title:       "IP 白名单",
			description: "启用后仅允许白名单中的 IP 访问",
			value:       boolToString(m.ipWhitelistEnabled),
			valueType:   "toggle",
			enabled:     m.ipWhitelistEnabled,
		},
		{
			key:         "tls_enabled",
			title:       "TLS 加密",
			description: "启用 HTTPS 加密传输",
			value:       boolToString(m.tlsEnabled),
			valueType:   "toggle",
			enabled:     m.tlsEnabled,
		},
		{
			key:         "auto_cert",
			title:       "自动证书",
			description: "自动获取和管理 SSL 证书 (Let's Encrypt)",
			value:       boolToString(m.autoCertEnabled),
			valueType:   "toggle",
			enabled:     m.autoCertEnabled,
		},
	}

	// 如果有 TLS 域名配置，显示域名
	if m.tlsDomains != "" {
		m.items = append(m.items, ConfigItem{
			key:         "tls_domains",
			title:       "证书域名",
			description: "自动证书管理的域名列表",
			value:       m.tlsDomains,
			valueType:   "text",
			enabled:     false,
		})
	}

	items := make([]list.Item, len(m.items))
	for i, item := range m.items {
		items[i] = item
	}
	m.list.SetItems(items)
}

// boolToString 将布尔值转换为字符串
func boolToString(b bool) string {
	if b {
		return "已启用"
	}
	return "已禁用"
}

// Update 处理消息
func (m *GlobalConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MsgGlobalConfigLoaded:
		m.ipWhitelistEnabled = msg.IPWhitelistEnabled
		m.tlsEnabled = msg.TLSEnabled
		m.autoCertEnabled = msg.AutoCertEnabled
		m.tlsDomains = msg.TLSDomains
		m.serverHost = msg.ServerHost
		m.serverPort = msg.ServerPort
		m.serverTLSPort = msg.ServerTLSPort
		m.proxyTimeouts = msg.ProxyTimeouts
		m.refreshList()

	case MsgRefresh:
		return m, m.loadConfig()

	case tea.KeyMsg:
		switch m.state {
		case StateList:
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.selectedItem > 0 {
					m.selectedItem--
				}
			case key.Matches(msg, m.keys.Down):
				if m.selectedItem < len(m.items)-1 {
					m.selectedItem++
				}
			case key.Matches(msg, m.keys.Toggle), key.Matches(msg, m.keys.Enter):
				if m.selectedItem < len(m.items) && m.items[m.selectedItem].valueType == "toggle" {
					return m, m.toggleConfig(m.items[m.selectedItem].key)
				}
			}
		}
	}

	// 更新列表
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// toggleConfig 切换配置开关
func (m *GlobalConfigModel) toggleConfig(key string) tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("数据库连接失败")}
		}

		ctx := context.Background()
		var err error

		switch key {
		case ConfigKeyIPWhitelistEnabled:
			newValue := "true"
			if m.ipWhitelistEnabled {
				newValue = "false"
			}
			err = queries.SetGlobalConfig(ctx, &db.SetGlobalConfigParams{
				Key:   key,
				Value: newValue,
			})

		case "tls_enabled":
			// 切换 TLS 开关
			tlsConfig, _ := queries.GetTLSConfig(ctx)
			newEnabled := sql.NullInt64{Int64: 0, Valid: true}
			if tlsConfig.Enabled.Int64 != 1 {
				newEnabled = sql.NullInt64{Int64: 1, Valid: true}
			}
			err = queries.UpsertTLSConfig(ctx, &db.UpsertTLSConfigParams{
				Enabled:   newEnabled,
				CertFile:  tlsConfig.CertFile,
				KeyFile:   tlsConfig.KeyFile,
				AutoCert:  tlsConfig.AutoCert,
				Domains:   tlsConfig.Domains,
			})

		case "auto_cert":
			// 切换自动证书开关
			tlsConfig, _ := queries.GetTLSConfig(ctx)
			newAutoCert := sql.NullInt64{Int64: 0, Valid: true}
			if tlsConfig.AutoCert.Int64 != 1 {
				newAutoCert = sql.NullInt64{Int64: 1, Valid: true}
			}
			err = queries.UpsertTLSConfig(ctx, &db.UpsertTLSConfigParams{
				Enabled:   tlsConfig.Enabled,
				CertFile:  tlsConfig.CertFile,
				KeyFile:   tlsConfig.KeyFile,
				AutoCert:  newAutoCert,
				Domains:   tlsConfig.Domains,
			})
		}

		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadConfig(), SendSuccess("配置已更新"))()
	}
}

// View 渲染模型
func (m *GlobalConfigModel) View() string {
	var b strings.Builder

	b.WriteString(styleSectionTitle.Render("▸ 系统配置"))
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(styleEmptyState.Render("  加载中..."))
	} else {
		for i, item := range m.items {
			selected := i == m.selectedItem
			prefix := "  "
			if selected {
				prefix = "▸ "
			}

			// 渲染配置项
			var valueStyle = styleItem
			if item.enabled {
				valueStyle = styleSuccess
			}

			line := prefix + item.title + ": "

			if selected {
				b.WriteString(styleItemSelected.Render(line))
				b.WriteString(valueStyle.Render(item.value))
			} else {
				b.WriteString(styleItem.Render(line))
				b.WriteString(valueStyle.Render(item.value))
			}
			b.WriteString("\n")

			// 渲染描述
			if item.description != "" {
				descPrefix := "    "
				b.WriteString(styleCardDesc.Render(descPrefix + item.description))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	// 帮助提示
	b.WriteString(Help("↑↓ 导航", "Space/Enter 切换开关", "q 退出"))

	return b.String()
}

// GetState 获取当前状态
func (m *GlobalConfigModel) GetState() AppState {
	return m.state
}