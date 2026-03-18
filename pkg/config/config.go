package config

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/repository"
)

// ValueType 定义配置值类型
type ValueType string

const (
	ValueTypeString ValueType = "string"
	ValueTypeInt    ValueType = "int"
	ValueTypeBool   ValueType = "bool"
)

// 默认配置值
var defaultConfig = map[string]struct {
	value     string
	valueType ValueType
}{
	"server.host":     {"0.0.0.0", ValueTypeString},
	"server.port":     {"8080", ValueTypeInt},
	"server.tls_port": {"8443", ValueTypeInt},
	"proxy.read_timeout":  {"30", ValueTypeInt},
	"proxy.write_timeout": {"30", ValueTypeInt},
	"proxy.idle_timeout":  {"120", ValueTypeInt},
}

// ConfigManager 配置管理器
type ConfigManager struct {
	queries *db.Queries
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		queries: repository.New(),
	}
}

// NewConfigManagerWithQueries 使用指定的 Queries 创建配置管理器
func NewConfigManagerWithQueries(queries *db.Queries) *ConfigManager {
	return &ConfigManager{
		queries: queries,
	}
}

// LoadAppConfig 从数据库加载所有应用配置并返回 *g.AppConfig
func (m *ConfigManager) LoadAppConfig(ctx context.Context) (*g.AppConfig, error) {
	config := &g.AppConfig{
		Server: g.ServerConfig{
			Host:    m.getStringWithDefault(ctx, "server.host", "0.0.0.0"),
			Port:    m.getIntWithDefault(ctx, "server.port", 8080),
			TLSPort: m.getIntWithDefault(ctx, "server.tls_port", 8443),
		},
		Proxy: g.ProxyConfig{
			ReadTimeout:  m.getIntWithDefault(ctx, "proxy.read_timeout", 30),
			WriteTimeout: m.getIntWithDefault(ctx, "proxy.write_timeout", 30),
			IdleTimeout:  m.getIntWithDefault(ctx, "proxy.idle_timeout", 120),
		},
	}
	return config, nil
}

// GetConfig 获取单个配置值（返回字符串形式）
func (m *ConfigManager) GetConfig(ctx context.Context, key string) (string, error) {
	appConfig, err := m.queries.GetAppConfig(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			// 尝试从默认配置获取
			if def, ok := defaultConfig[key]; ok {
				return def.value, nil
			}
			return "", fmt.Errorf("配置项 %s 不存在", key)
		}
		return "", fmt.Errorf("获取配置失败: %w", err)
	}
	return appConfig.Value, nil
}

// GetConfigWithType 获取配置值并按指定类型返回
func (m *ConfigManager) GetConfigWithType(ctx context.Context, key string, valueType ValueType) (interface{}, error) {
	value, err := m.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}

	switch valueType {
	case ValueTypeInt:
		return strconv.Atoi(value)
	case ValueTypeBool:
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}

// SetConfig 设置配置值（字符串类型）
func (m *ConfigManager) SetConfig(ctx context.Context, key, value string) error {
	return m.SetConfigWithMeta(ctx, key, value, ValueTypeString, "")
}

// SetConfigWithMeta 设置配置值（带类型和描述）
func (m *ConfigManager) SetConfigWithMeta(ctx context.Context, key, value string, valueType ValueType, description string) error {
	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	err := m.queries.SetAppConfig(ctx, &db.SetAppConfigParams{
		Key:         key,
		Value:       value,
		ValueType:   string(valueType),
		Description: desc,
	})
	if err != nil {
		return fmt.Errorf("设置配置失败: %w", err)
	}
	return nil
}

// SetIntConfig 设置整数类型配置
func (m *ConfigManager) SetIntConfig(ctx context.Context, key string, value int, description string) error {
	return m.SetConfigWithMeta(ctx, key, strconv.Itoa(value), ValueTypeInt, description)
}

// SetBoolConfig 设置布尔类型配置
func (m *ConfigManager) SetBoolConfig(ctx context.Context, key string, value bool, description string) error {
	return m.SetConfigWithMeta(ctx, key, strconv.FormatBool(value), ValueTypeBool, description)
}

// ListConfigs 列出所有配置
func (m *ConfigManager) ListConfigs(ctx context.Context) ([]*db.AppConfig, error) {
	return m.queries.ListAppConfig(ctx)
}

// DeleteConfig 删除配置
func (m *ConfigManager) DeleteConfig(ctx context.Context, key string) error {
	return m.queries.DeleteAppConfig(ctx, key)
}

// GetIntConfig 获取整数类型配置
func (m *ConfigManager) GetIntConfig(ctx context.Context, key string) (int, error) {
	value, err := m.GetConfig(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// GetBoolConfig 获取布尔类型配置
func (m *ConfigManager) GetBoolConfig(ctx context.Context, key string) (bool, error) {
	value, err := m.GetConfig(ctx, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value)
}

// getStringWithDefault 获取字符串配置，带默认值
func (m *ConfigManager) getStringWithDefault(ctx context.Context, key, defaultValue string) string {
	value, err := m.GetConfig(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// getIntWithDefault 获取整数配置，带默认值
func (m *ConfigManager) getIntWithDefault(ctx context.Context, key string, defaultValue int) int {
	value, err := m.GetIntConfig(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// getBoolWithDefault 获取布尔配置，带默认值
func (m *ConfigManager) getBoolWithDefault(ctx context.Context, key string, defaultValue bool) bool {
	value, err := m.GetBoolConfig(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}