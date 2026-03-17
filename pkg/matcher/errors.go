package matcher

import "errors"

// 匹配器错误定义
var (
	// ErrNilRule 规则为空
	ErrNilRule = errors.New("rule is nil")
	// ErrEmptyPathPattern 路径模式为空
	ErrEmptyPathPattern = errors.New("path pattern is empty")
	// ErrEmptyCelExpression CEL表达式为空
	ErrEmptyCelExpression = errors.New("CEL expression is empty")
	// ErrEmptyPluginName 插件名称为空
	ErrEmptyPluginName = errors.New("plugin name is empty")
	// ErrNilPluginManager 插件管理器为空
	ErrNilPluginManager = errors.New("plugin manager is nil")
	// ErrPluginNotFound 插件未找到
	ErrPluginNotFound = errors.New("plugin not found")
	// ErrNilCelEngine CEL引擎为空
	ErrNilCelEngine = errors.New("CEL engine is nil")
)