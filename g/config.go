package g

// DefaultDbPath 默认数据库文件路径
const DefaultDbPath = "prism.db"

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
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	TLSPort int    `mapstructure:"tls_port"`
}

// ProxyConfig contains proxy configuration.
type ProxyConfig struct {
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	IdleTimeout  int `mapstructure:"idle_timeout"`
}