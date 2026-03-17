package matcher

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// extractPathParams 从路径中提取参数
// pattern 格式：/users/{id}/orders/{orderId}
// path 格式：/users/123/orders/456
// 返回：map[string]string{"id": "123", "orderId": "456"}
func extractPathParams(path, pattern string) map[string]string {
	params := make(map[string]string)

	// 分割路径和模式
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	// 长度必须相等
	if len(pathParts) != len(patternParts) {
		return params
	}

	// 匹配参数占位符的正则表达式
	paramRegex := regexp.MustCompile(`^\{([a-zA-Z_][a-zA-Z0-9_]*)\}$`)

	for i, part := range patternParts {
		// 检查是否是参数占位符
		matches := paramRegex.FindStringSubmatch(part)
		if len(matches) == 2 {
			// 提取参数名和值
			paramName := matches[1]
			paramValue := pathParts[i]
			params[paramName] = paramValue
		}
	}

	return params
}

// matchPathPattern 检查路径是否匹配模式
// pattern 格式：/users/{id}/orders/{orderId}
// path 格式：/users/123/orders/456
func matchPathPattern(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	// 长度必须相等
	if len(pathParts) != len(patternParts) {
		return false
	}

	// 匹配参数占位符的正则表达式
	paramRegex := regexp.MustCompile(`^\{[a-zA-Z_][a-zA-Z0-9_]*\}$`)

	for i, part := range patternParts {
		// 如果是参数占位符，跳过比较
		if paramRegex.MatchString(part) {
			continue
		}
		// 否则必须精确匹配
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

// getURLParams 获取URL查询参数
func getURLParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	query := c.Request.URL.Query()
	for key, values := range query {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// getHeaders 获取请求头
func getHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// getBody 获取并解析JSON请求体
// 注意：这会消耗请求体，需要确保在调用前已缓存请求体
func getBody(c *gin.Context) map[string]any {
	body := make(map[string]any)

	// 读取请求体
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return body
	}

	// 恢复请求体供后续使用
	c.Request.Body = io.NopCloser(strings.NewReader(string(data)))

	// 如果请求体为空，返回空map
	if len(data) == 0 {
		return body
	}

	// 解析JSON
	if err := json.Unmarshal(data, &body); err != nil {
		// 如果不是JSON，尝试解析为数组或其他类型
		return body
	}

	return body
}

// getFormValues 获取表单数据
func getFormValues(c *gin.Context) map[string]string {
	form := make(map[string]string)

	// 解析表单数据
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		// 尝试解析普通表单
		if err := c.Request.ParseForm(); err != nil {
			return form
		}
	}

	// 合并表单值
	for key, values := range c.Request.Form {
		if len(values) > 0 {
			form[key] = values[0]
		}
	}

	// 合并multipart表单值
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			if len(values) > 0 {
				form[key] = values[0]
			}
		}
	}

	return form
}

// mergeParams 合并多个参数map
func mergeParams(paramsMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, params := range paramsMaps {
		for k, v := range params {
			result[k] = v
		}
	}
	return result
}

// convertPatternToGin 将路径模式转换为Gin路由格式
// /users/{id}/orders/{orderId} -> /users/:id/orders/:orderId
func convertPatternToGin(pattern string) string {
	// 替换 {param} 为 :param
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	return re.ReplaceAllString(pattern, ":$1")
}

// convertPatternFromGin 将Gin路由格式转换为路径模式
// /users/:id/orders/:orderId -> /users/{id}/orders/{orderId}
func convertPatternFromGin(ginPattern string) string {
	// 替换 :param 为 {param}
	re := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	return re.ReplaceAllString(ginPattern, `{$1}`)
}