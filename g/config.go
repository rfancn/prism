package g

// Config is the global configuration instance.
var Config RootConfig

// RootConfig is the root configuration structure.
type RootConfig struct {
	Sdk SdkConfig `mapstructure:"sdk"`
	App AppConfig `mapstructure:"app"`
}

// SdkConfig contains SDK-level configurations.
type SdkConfig struct {
	Log    LogConfig    `mapstructure:"log"`
	Sqlite SqliteConfig `mapstructure:"sqlite"`
}

// LogConfig contains logging configuration.
type LogConfig struct {
	Level    string       `mapstructure:"level"`
	Filename string       `mapstructure:"filename"`
	Rotate   RotateConfig `mapstructure:"rotate"`
}

// RotateConfig contains log rotation configuration.
type RotateConfig struct {
	MaxAge       int `mapstructure:"max_age"`
	RotationTime int `mapstructure:"rotation_time"`
}

// SqliteConfig contains SQLite database configuration.
type SqliteConfig struct {
	Db string `mapstructure:"db"`
}

// AppConfig contains application-level configurations.
type AppConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Proxy     ProxyConfig     `mapstructure:"proxy"`
	Route     RouteConfig     `mapstructure:"route"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
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

// RouteConfig contains route configuration.
type RouteConfig struct {
	ProtectPrefix string `mapstructure:"protect_prefix"`
	PublicPrefix  string `mapstructure:"public_prefix"`
}

// RateLimitConfig contains rate limiting configuration.
type RateLimitConfig struct {
	WindowSize int `mapstructure:"window_size"`
	Limit      int `mapstructure:"limit"`
}