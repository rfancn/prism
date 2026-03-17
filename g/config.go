package g

// Config is the global configuration instance.
var Config RootConfig

// RootConfig is the root configuration structure.
type RootConfig struct {
	App AppConfig `mapstructure:"app"`
}

// AppConfig contains application-level configurations.
type AppConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Proxy  ProxyConfig  `mapstructure:"proxy"`
	Plugin PluginConfig `mapstructure:"plugin"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	TLSPort int `mapstructure:"tls_port"`
}

// ProxyConfig contains proxy configuration.
type ProxyConfig struct {
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	IdleTimeout  int `mapstructure:"idle_timeout"`
}

// PluginConfig contains plugin configuration.
type PluginConfig struct {
	// Paths 插件搜索路径列表
	Paths []string `mapstructure:"paths"`
}
