package matcher

import (
	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
)

// URLParamMatcher URL参数匹配器
// 用于从URL查询参数中提取值并执行CEL表达式进行条件判断
type URLParamMatcher struct {
	celEngine *cel.Engine
}

// NewURLParamMatcher 创建URL参数匹配器
func NewURLParamMatcher(celEngine *cel.Engine) *URLParamMatcher {
	return &URLParamMatcher{
		celEngine: celEngine,
	}
}

// Match 执行URL参数匹配
func (m *URLParamMatcher) Match(c *gin.Context, rule *db.RouteRule) Result {
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

	// 3. 获取URL查询参数
	urlParams := getURLParams(c)

	// 4. 准备CEL上下文
	ctx := &cel.MatchContext{
		PathParams: make(map[string]string), // URL参数匹配器不使用路径参数
		URLParams:  urlParams,
		Headers:    getHeaders(c),
		Body:       getBody(c),
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
	// URL参数匹配器返回所有URL参数作为提取的参数
	return Result{
		Matched: matched,
		Params:  urlParams,
		Error:   nil,
	}
}