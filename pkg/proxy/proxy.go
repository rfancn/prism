// Package proxy provides HTTP reverse proxy functionality.
package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/pkg/types"
)

// ProxyHandler handles HTTP request proxying.
type ProxyHandler struct {
	targetTLS *types.TargetTLSConfig
}

// NewProxyHandler creates a new proxy handler.
func NewProxyHandler(targetTLS *types.TargetTLSConfig) *ProxyHandler {
	return &ProxyHandler{
		targetTLS: targetTLS,
	}
}

// ForwardOptions 转发选项
type ForwardOptions struct {
	// TargetURL 目标URL
	TargetURL string
	// Params 从请求中提取的参数（用于URL替换）
	Params map[string]string
	// SourceName 来源名称（用于路径处理）
	SourceName string
	// ExtraHeaders 额外的请求头信息
	ExtraHeaders map[string]string
}

// Forward 转发请求到目标服务器
func (p *ProxyHandler) Forward(c *gin.Context, opts *ForwardOptions) error {
	// 解析目标URL
	target, err := url.Parse(opts.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	sdk.Logger().Debug("forwarding request",
		"target", opts.TargetURL,
		"path", c.Request.URL.Path,
		"params", opts.Params,
	)

	// 创建反向代理
	proxy := &httputil.ReverseProxy{
		Director:    p.createDirector(target, c, opts),
		Transport:   createTransport(p.targetTLS),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			sdk.Logger().Error("proxy error", "err", err, "target", opts.TargetURL)
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(gin.H{"error": "bad gateway"})
		},
	}

	// 转发请求
	proxy.ServeHTTP(c.Writer, c.Request)
	return nil
}

// createDirector creates a director function for the reverse proxy.
func (p *ProxyHandler) createDirector(target *url.URL, c *gin.Context, opts *ForwardOptions) func(*http.Request) {
	return func(req *http.Request) {
		// Set the target URL
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// Build the path: target path + remaining path after route match
		targetPath := target.Path
		if targetPath == "" {
			targetPath = "/"
		}

		// Get the remaining path
		remainingPath := c.Request.URL.Path

		// 如果有来源名称，去除来源前缀
		if opts != nil && opts.SourceName != "" {
			sourcePrefix := "/" + opts.SourceName
			if strings.HasPrefix(remainingPath, sourcePrefix) {
				remainingPath = strings.TrimPrefix(remainingPath, sourcePrefix)
			}
		}

		// 组合目标路径和剩余路径
		if strings.HasSuffix(targetPath, "/") {
			req.URL.Path = targetPath + strings.TrimPrefix(remainingPath, "/")
		} else {
			if remainingPath != "" && remainingPath != "/" {
				req.URL.Path = targetPath + remainingPath
			} else {
				req.URL.Path = targetPath
			}
		}

		// Inject extra headers
		if opts != nil && opts.ExtraHeaders != nil {
			for key, value := range opts.ExtraHeaders {
				req.Header.Set(key, value)
			}
		}

		// Set X-Forwarded headers
		if req.Header.Get("X-Forwarded-For") == "" {
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		}
		if req.Header.Get("X-Forwarded-Host") == "" {
			req.Header.Set("X-Forwarded-Host", req.Host)
		}
		if req.Header.Get("X-Forwarded-Proto") == "" {
			if req.TLS != nil {
				req.Header.Set("X-Forwarded-Proto", "https")
			} else {
				req.Header.Set("X-Forwarded-Proto", "http")
			}
		}
	}
}