package router

import (
	"context"

	"github.com/hdget/sdk"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/repository"
)

// Loader 路由加载器
// 负责从数据库加载路由配置并缓存
type Loader struct {
	queries *db.Queries
}

// NewLoader 创建路由加载器
func NewLoader() *Loader {
	return &Loader{
		queries: repository.New(),
	}
}

// LoadAll 加载所有路由配置
func (l *Loader) LoadAll(ctx context.Context) (*RouterConfig, error) {
	// 加载所有启用的来源
	sources, err := l.queries.ListEnabledSources(ctx)
	if err != nil {
		return nil, err
	}

	config := &RouterConfig{
		Sources: make([]*SourceConfig, 0, len(sources)),
	}

	for _, source := range sources {
		sourceConfig, err := l.loadSource(ctx, source)
		if err != nil {
			sdk.Logger().Error("failed to load source", "source_id", source.ID, "err", err)
			continue
		}
		config.Sources = append(config.Sources, sourceConfig)
	}

	sdk.Logger().Info("router config loaded",
		"sources_count", len(config.Sources),
	)

	return config, nil
}

// loadSource 加载来源配置
func (l *Loader) loadSource(ctx context.Context, source *db.Source) (*SourceConfig, error) {
	// 加载该来源下所有启用的项目
	projects, err := l.queries.ListEnabledProjectsBySourceID(ctx, source.ID)
	if err != nil {
		return nil, err
	}

	sourceConfig := &SourceConfig{
		Source:   source,
		Projects: make([]*ProjectConfig, 0, len(projects)),
	}

	for _, project := range projects {
		projectConfig, err := l.loadProject(ctx, project)
		if err != nil {
			sdk.Logger().Error("failed to load project", "project_id", project.ID, "err", err)
			continue
		}
		sourceConfig.Projects = append(sourceConfig.Projects, projectConfig)
	}

	return sourceConfig, nil
}

// loadProject 加载项目配置
func (l *Loader) loadProject(ctx context.Context, project *db.Project) (*ProjectConfig, error) {
	// 加载该项目下所有启用的路由规则
	rules, err := l.queries.ListEnabledRouteRulesByProjectID(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	projectConfig := &ProjectConfig{
		Project: project,
		Rules:   make([]*RouteRuleConfig, 0, len(rules)),
	}

	for _, rule := range rules {
		ruleConfig, err := l.loadRule(ctx, rule)
		if err != nil {
			sdk.Logger().Error("failed to load rule", "rule_id", rule.ID, "err", err)
			continue
		}
		projectConfig.Rules = append(projectConfig.Rules, ruleConfig)
	}

	return projectConfig, nil
}

// loadRule 加载路由规则配置
func (l *Loader) loadRule(ctx context.Context, rule *db.RouteRule) (*RouteRuleConfig, error) {
	return &RouteRuleConfig{
		Rule: rule,
	}, nil
}

// GetSourceByName 根据名称获取来源
func (l *Loader) GetSourceByName(ctx context.Context, name string) (*db.Source, error) {
	return l.queries.GetSourceByName(ctx, name)
}

// GetProjectsBySourceID 根据来源ID获取项目列表
func (l *Loader) GetProjectsBySourceID(ctx context.Context, sourceID string) ([]*db.Project, error) {
	return l.queries.ListEnabledProjectsBySourceID(ctx, sourceID)
}

// GetRulesByProjectID 根据项目ID获取路由规则列表
func (l *Loader) GetRulesByProjectID(ctx context.Context, projectID string) ([]*db.RouteRule, error) {
	return l.queries.ListEnabledRouteRulesByProjectID(ctx, projectID)
}