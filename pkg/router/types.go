// Package router 提供路由管理功能
// 实现三层路由结构：Source -> Project -> RouteRule
package router

import (
	"github.com/rfancn/prism/autogen/db"
)

// MatchContext 匹配上下文
// 用于存储请求匹配过程中提取的信息
type MatchContext struct {
	// Source 匹配的来源
	Source *db.Source
	// Project 匹配的项目
	Project *db.Project
	// Rule 匹配的路由规则
	Rule *db.RouteRule
	// Params 从请求中提取的参数
	Params map[string]string
}

// SourceConfig 来源配置
// 缓存来源及其关联的项目和路由规则
type SourceConfig struct {
	Source   *db.Source
	Projects []*ProjectConfig
}

// ProjectConfig 项目配置
// 缓存项目及其关联的路由规则
type ProjectConfig struct {
	Project *db.Project
	Rules   []*RouteRuleConfig
}

// RouteRuleConfig 路由规则配置
type RouteRuleConfig struct {
	Rule *db.RouteRule
}

// RouterConfig 路由器完整配置
// 包含所有来源配置的缓存
type RouterConfig struct {
	Sources []*SourceConfig
}

// ForwardOptions 转发选项
type ForwardOptions struct {
	// TargetURL 目标URL（可能包含参数占位符）
	TargetURL string
	// Params 从请求中提取的参数（用于URL替换）
	Params map[string]string
}