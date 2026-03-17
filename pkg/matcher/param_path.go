package matcher

import (
	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
)

// ParamPathMatcher 参数化路径匹配器
// 用于从路径中提取参数并执行CEL表达式进行条件判断
type ParamPathMatcher struct {
	celEngine *cel.Engine
}

// NewParamPathMatcher 创建参数化路径匹配器
func NewParamPathMatcher(celEngine *cel.Engine) *ParamPathMatcher {
	return &ParamPathMatcher{
		celEngine: celEngine,
	}
}

// Match 执行参数化路径匹配
func (m *ParamPathMatcher) Match(c *gin.Context, rule *db.RouteRule) Result {
	// 1. 检查规则是否有效
	if rule == nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrNilRule,
		}
	}

	// 2. 获取路径模式
	pathPattern := rule.PathPattern
	if !pathPattern.Valid || pathPattern.String == "" {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   ErrEmptyPathPattern,
		}
	}

	// 3. 获取当前请求路径
	requestPath := c.Request.URL.Path

	// 4. 检查路径是否匹配模式
	if !matchPathPattern(requestPath, pathPattern.String) {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   nil,
		}
	}

	// 5. 提取路径参数
	pathParams := extractPathParams(requestPath, pathPattern.String)

	// 6. 如果没有CEL表达式，直接返回匹配成功
	celExpr := rule.CelExpression
	if !celExpr.Valid || celExpr.String == "" {
		return Result{
			Matched: true,
			Params:  pathParams,
			Error:   nil,
		}
	}

	// 7. 准备CEL上下文
	ctx := &cel.MatchContext{
		PathParams: pathParams,
		URLParams:  getURLParams(c),
		Headers:    getHeaders(c),
		Body:       getBody(c),
	}

	// 8. 执行CEL表达式
	matched, _, err := m.celEngine.Evaluate(celExpr.String, ctx)
	if err != nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   err,
		}
	}

	// 9. 返回匹配结果
	return Result{
		Matched: matched,
		Params:  pathParams,
		Error:   nil,
	}
}