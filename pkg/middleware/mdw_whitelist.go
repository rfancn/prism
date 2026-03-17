package middleware

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/repository"
)

// 全局配置键名
const GlobalConfigKeyWhitelistEnabled = "ip_whitelist_enabled"

// WhitelistMiddleware checks if the client IP is in the whitelist.
func NewWhitelistMiddleware() (gin.HandlerFunc, error) {
	return func(c *gin.Context) {
		queries := repository.New()
		if queries == nil {
			c.Next()
			return
		}

		// 检查全局开关是否启用
		config, err := queries.GetGlobalConfig(context.Background(), GlobalConfigKeyWhitelistEnabled)
		if err != nil {
			// 如果配置不存在，默认关闭IP白名单功能
			c.Next()
			return
		}

		// 如果值为 "false" 或 "0"，则关闭IP白名单检查
		if config.Value != "true" && config.Value != "1" {
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		// Get all enabled whitelist entries
		entries, err := queries.ListEnabledWhitelist(context.Background())
		if err != nil {
			sdk.Logger().Error("failed to get whitelist", "err", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		// If no whitelist entries, allow all
		if len(entries) == 0 {
			c.Next()
			return
		}

		// Check if IP is whitelisted
		allowed := false
		for _, entry := range entries {
			if isIPAllowed(clientIP, entry.IpCidr) {
				allowed = true
				break
			}
		}

		if !allowed {
			sdk.Logger().Info("IP not whitelisted", "ip", clientIP)
			c.JSON(403, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		c.Next()
	}, nil
}

// isIPAllowed checks if an IP is allowed by a CIDR or exact match.
func isIPAllowed(ipStr, cidrOrIP string) bool {
	// Parse the client IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Try CIDR parsing first
	_, ipNet, err := net.ParseCIDR(cidrOrIP)
	if err == nil {
		return ipNet.Contains(ip)
	}

	// Try exact IP match
	allowedIP := net.ParseIP(cidrOrIP)
	if allowedIP != nil {
		return ip.Equal(allowedIP)
	}

	return false
}