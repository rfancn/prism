package matcher

import (
	"github.com/gin-gonic/gin"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/pkg/cel"
)

// RequestFormMatcher 表单数据匹配器
// 用于解析表单数据并执行CEL表达式进行条件判断
type RequestFormMatcher struct {
	celEngine *cel.Engine
}

// NewRequestFormMatcher 创建表单数据匹配器
func NewRequestFormMatcher(celEngine *cel.Engine) *RequestFormMatcher {
	return &RequestFormMatcher{
		celEngine: celEngine,
	}
}

// Match 执行表单数据匹配
func (m *RequestFormMatcher) Match(c *gin.Context, rule *db.RouteRule) Result {
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

	// 3. 获取表单数据
	formValues := getFormValues(c)

	// 4. 将表单数据转换为body格式供CEL使用
	body := formToBody(formValues)

	// 5. 准备CEL上下文
	ctx := &cel.MatchContext{
		PathParams: make(map[string]string),
		URLParams:  getURLParams(c),
		Headers:    getHeaders(c),
		Body:       body,
	}

	// 6. 执行CEL表达式
	matched, _, err := m.celEngine.Evaluate(celExpr.String, ctx)
	if err != nil {
		return Result{
			Matched: false,
			Params:  nil,
			Error:   err,
		}
	}

	// 7. 返回匹配结果
	return Result{
		Matched: matched,
		Params:  formValues,
		Error:   nil,
	}
}

// formToBody 将表单数据转换为body格式
// 用于CEL表达式访问
func formToBody(form map[string]string) map[string]any {
	body := make(map[string]any)
	for k, v := range form {
		body[k] = v
	}
	return body
}