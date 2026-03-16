package server

import (
	"golang.org/x/crypto/acme/autocert"
)

// DefaultCertCacheDir 默认证书缓存目录
const DefaultCertCacheDir = "./certs"

// NewAutoCertManager 创建自动证书管理器
// domains: 域名列表
// cacheDir: 证书缓存目录，为空则使用默认目录
func NewAutoCertManager(domains []string, cacheDir string) *autocert.Manager {
	if cacheDir == "" {
		cacheDir = DefaultCertCacheDir
	}
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(cacheDir),
	}
}