package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/repository"
)

// ContextKeyUserID is the context key for user ID.
const ContextKeyUserID = "userID"

// AuthMiddleware validates API key from request.
func NewAuthMiddleware() (gin.HandlerFunc, error) {
	return func(c *gin.Context) {
		queries := repository.New()
		if queries == nil {
			c.Next()
			return
		}

		// Get API key from header or query
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			sdk.Logger().Debug("missing API key")
			c.JSON(401, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Validate API key
		keyInfo, err := queries.GetAPIKeyByKeyEnabled(context.Background(), apiKey)
		if err != nil {
			sdk.Logger().Debug("invalid API key", "err", err)
			c.JSON(401, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Update last used time
		go func() {
			_ = queries.UpdateAPIKeyLastUsed(context.Background(), keyInfo.ID)
		}()

		// Store user ID in context for rate limiting
		c.Set(ContextKeyUserID, keyInfo.UserID)

		c.Next()
	}, nil
}

// GetUserID retrieves the user ID from context.
func GetUserID(c *gin.Context) string {
	if v, exists := c.Get(ContextKeyUserID); exists {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}