// Package proxy provides HTTP reverse proxy functionality.
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/parser"
	"github.com/rfancn/prism/pkg/types"
	"github.com/rfancn/prism/repository"
)

// ProxyHandler handles HTTP request proxying.
type ProxyHandler struct {
	jsonParser *parser.JSONBodyParser
	targetTLS  *types.TargetTLSConfig
}

// NewProxyHandler creates a new proxy handler.
func NewProxyHandler(targetTLS *types.TargetTLSConfig) *ProxyHandler {
	return &ProxyHandler{
		targetTLS: targetTLS,
	}
}

// Handler returns a Gin handler function for proxying requests.
func (p *ProxyHandler) Handler(route *db.Route) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Extract identifier based on source type
		identifier, err := p.extractIdentifier(c, route)
		if err != nil {
			sdk.Logger().Debug("failed to extract identifier", "err", err, "source", route.IdentifierSource)
		}

		sdk.Logger().Debug("processing request",
			"route_id", route.ID,
			"identifier", identifier,
			"target", route.TargetUrl,
			"path", c.Request.URL.Path,
		)

		// Get headers for the route
		var headers []*db.Header
		queries := repository.New()
		if queries != nil {
			headers, err = queries.GetHeadersByRouteID(ctx, route.ID)
			if err != nil {
				sdk.Logger().Error("failed to get headers", "err", err)
			}
		}

		// Create reverse proxy
		targetURL, err := url.Parse(route.TargetUrl)
		if err != nil {
			p.respondError(c, 500, fmt.Sprintf("invalid target URL: %v", err))
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: p.createDirector(targetURL, headers, c),
			Transport: createTransport(p.targetTLS),
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				sdk.Logger().Error("proxy error", "err", err)
				w.WriteHeader(502)
				json.NewEncoder(w).Encode(gin.H{"error": "bad gateway"})
			},
		}

		// Forward the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// extractIdentifier extracts the identifier from the request based on the route configuration.
func (p *ProxyHandler) extractIdentifier(c *gin.Context, route *db.Route) (string, error) {
	switch route.IdentifierSource {
	case "path":
		// Use Gin's built-in path parameter extraction
		return c.Param(route.Identifier), nil

	case "url_param":
		// Use Gin's built-in query parameter extraction
		value := c.Query(route.Identifier)
		if value == "" {
			return "", fmt.Errorf("query parameter %s not found", route.Identifier)
		}
		return value, nil

	case "json_body":
		// JSON body extraction requires custom parsing
		return p.extractFromJSONBody(c, route.Identifier)

	default:
		return "", fmt.Errorf("unknown identifier source: %s", route.IdentifierSource)
	}
}

// extractFromJSONBody extracts a field from the JSON request body.
func (p *ProxyHandler) extractFromJSONBody(c *gin.Context, fieldName string) (string, error) {
	// Read body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	// Restore body for downstream use
	c.Request.Body = io.NopCloser(NewBufferedReader(bodyBytes))

	// Parse JSON and extract field
	if p.jsonParser == nil {
		p.jsonParser = parser.NewJSONBodyParser(fieldName)
	} else {
		// Update field name if different
		p.jsonParser = parser.NewJSONBodyParser(fieldName)
	}

	return p.jsonParser.Extract(bodyBytes)
}

// respondError sends an error response.
func (p *ProxyHandler) respondError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// BufferedReader is a simple reader that can be reset.
type BufferedReader struct {
	data []byte
	pos  int
}

// NewBufferedReader creates a new buffered reader.
func NewBufferedReader(data []byte) *BufferedReader {
	return &BufferedReader{data: data}
}

func (r *BufferedReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// createDirector creates a director function for the reverse proxy.
func (p *ProxyHandler) createDirector(target *url.URL, headers []*db.Header, c *gin.Context) func(*http.Request) {
	return func(req *http.Request) {
		// Set the target URL
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// Build the path: target path + remaining path after route match
		targetPath := target.Path
		if targetPath == "" {
			targetPath = "/"
		}

		// Get the remaining path after the route pattern match
		remainingPath := c.Request.URL.Path
		if strings.HasSuffix(targetPath, "/") {
			req.URL.Path = targetPath + strings.TrimPrefix(remainingPath, "/")
		} else {
			req.URL.Path = targetPath + remainingPath
		}

		// Inject headers
		for _, h := range headers {
			req.Header.Set(h.Key, h.Value)
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