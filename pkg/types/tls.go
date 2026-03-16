// Package types provides shared types across packages.
package types

// TargetTLSConfig contains TLS configuration for target connections.
type TargetTLSConfig struct {
	InsecureSkipVerify bool
}

// ServerTLSConfig contains TLS configuration for the HTTP server.
type ServerTLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// AutoCertConfig 包含自动证书配置
type AutoCertConfig struct {
	Enabled  bool
	Domains  []string
	CacheDir string
}