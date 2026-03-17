package matcher

import (
	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
)

// RequestBodyMatcher 请求体匹配器
// 用于解析JSON请求体并执行CEL表达式进行条件判断
type RequestBodyMatcher struct {
	celEngine *cel.Engine
}

// NewRequestBodyMatcher 创建请求体匹配器
func NewRequestBodyMatcher(celEngine *cel.Engine) *RequestBodyMatcher {
	return &RequestBodyMatcher{
		celEngine: celEngine,
	}
}

// Match 执行请求体匹配
func (m *RequestBodyMatcher) Match(c *gin.Context, rule *db.RouteRule) Result {
	// 1. 检查规则是否有效
	if rule == nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrNilRule,
		}
	}

	// 2. 获取CEL表达式（必需）
	celExpr := rule.CelExpression
	if !celExpr.Valid || celExpr.String == "" {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrEmptyCelExpression,
		}
	}

	// 3. 获取请求体
	body := getBody(c)

	// 4. 准备CEL上下文
	ctx := &cel.MatchContext{
		PathParams: make(map[string]string),
		URLParams:  getURLParams(c),
		Headers:    getHeaders(c),
		Body:       body,
	}

	// 5. 执行CEL表达式
	matched, _, err := m.celEngine.Evaluate(celExpr.String, ctx)
	if err != nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   err,
		}
	}

	// 6. 返回匹配结果
	// 从请求体中提取字符串参数
	params := extractParamsFromBody(body)

	return Result{
		Matched: matched,
		Params:  params,
		Error:   nil,
	}
}

// extractParamsFromBody 从请求体中提取字符串类型的参数
func extractParamsFromBody(body map[string]any) map[string]string {
	params := make(map[string]string)
	for k, v := range body {
		// 只提取字符串类型的值
		if str, ok := v.(string); ok {
			params[k] = str
		}
	}
	return params
}