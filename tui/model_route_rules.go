package tui

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/repository"
)

// 匹配类型常量
const (
	MatchTypeParamPath   = "param_path"
	MatchTypeURLParam    = "url_param"
	MatchTypeRequestBody = "request_body"
	MatchTypeRequestForm = "request_form"
	MatchTypePlugin      = "plugin"
)

// 匹配类型标签
var matchTypeLabels = map[string]string{
	MatchTypeParamPath:   "路径参数",
	MatchTypeURLParam:    "URL参数",
	MatchTypeRequestBody: "请求体",
	MatchTypeRequestForm: "表单数据",
	MatchTypePlugin:      "插件处理",
}

// RouteRulesModel 管理路由规则列表
type RouteRulesModel struct {
	list              list.Model
	rules             []*db.RouteRule
	sources           []*db.Source         // 来源列表，用于选择器
	projects          []*db.Project        // 项目列表，用于选择器
	plugins           []*db.PluginRegistry // 插件列表
	state             AppState
	form              *Form
	selected          *db.RouteRule
	selectedProj      *db.Project // 当前筛选的项目
	contextSourceName string      // 上下文：来源名称（用于显示）
	contextProjName   string      // 上下文：项目名称（用于显示）
	width             int
	height            int
	keys              KeyMap
}

// NewRouteRulesModel 创建一个新的路由规则模型
func NewRouteRulesModel() *RouteRulesModel {
	m := &RouteRulesModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "路由规则列表", 80, 20)
	return m
}

// Init 初始化模型
func (m *RouteRulesModel) Init() tea.Cmd {
	return tea.Batch(m.loadSources(), m.loadProjects(), m.loadPlugins(), m.loadRules())
}

// SetSize 设置尺寸
func (m *RouteRulesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
	// 如果表单存在，也设置表单尺寸
	if m.form != nil {
		m.form.SetSize(width, height)
	}
}

// MsgProjectsLoadedForRule 项目加载完成消息（路由规则用）
type MsgProjectsLoadedForRule struct {
	Projects []*db.Project
}

// MsgSourcesLoadedForRule 来源加载完成消息（路由规则用）
type MsgSourcesLoadedForRule struct {
	Sources []*db.Source
}

// loadSources 加载来源列表
func (m *RouteRulesModel) loadSources() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		sources, err := queries.ListSources(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgSourcesLoadedForRule{Sources: sources}
	}
}

// MsgPluginsLoaded 插件加载完成消息
type MsgPluginsLoaded struct {
	Plugins []*db.PluginRegistry
}

// MsgRouteRulesLoaded 路由规则加载完成消息
type MsgRouteRulesLoaded struct {
	Rules []*db.RouteRule
}

// loadProjects 加载项目列表
func (m *RouteRulesModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		projects, err := queries.ListProjects(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgProjectsLoadedForRule{Projects: projects}
	}
}

// loadPlugins 加载插件列表
func (m *RouteRulesModel) loadPlugins() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		plugins, err := queries.ListPlugins(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgPluginsLoaded{Plugins: plugins}
	}
}

// loadRules 从数据库加载路由规则
func (m *RouteRulesModel) loadRules() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}

		var rules []*db.RouteRule
		var err error

		if m.selectedProj != nil {
			rules, err = queries.ListRouteRulesByProjectID(context.Background(), m.selectedProj.ID)
		} else {
			rules, err = queries.ListRouteRules(context.Background())
		}

		if err != nil {
			return MsgError{Err: err}
		}
		return MsgRouteRulesLoaded{Rules: rules}
	}
}

// refreshList 刷新列表
func (m *RouteRulesModel) refreshList() {
	items := make([]list.Item, len(m.rules))
	for i, r := range m.rules {
		// 获取项目名称和目标URL
		projectName := "未知项目"
		targetURL := ""
		for _, p := range m.projects {
			if p.ID == r.ProjectID {
				projectName = p.Name
				targetURL = p.TargetUrl.String
				break
			}
		}
		matchLabel := matchTypeLabels[r.MatchType]
		if matchLabel == "" {
			matchLabel = r.MatchType
		}
		items[i] = MenuItem{
			title:       fmt.Sprintf("%s/%s", projectName, r.Name),
			description: fmt.Sprintf("[%s] %s", matchLabel, Truncate(targetURL, 40)),
		}
	}
	m.list.SetItems(items)
}

// Update 处理消息
func (m *RouteRulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch typedMsg := msg.(type) {
	case MsgSourcesLoadedForRule:
		m.sources = typedMsg.Sources
		m.refreshList()

	case MsgProjectsLoadedForRule:
		m.projects = typedMsg.Projects
		m.refreshList()

	case MsgPluginsLoaded:
		m.plugins = typedMsg.Plugins

	case MsgRouteRulesLoaded:
		m.rules = typedMsg.Rules
		m.refreshList()

	case MsgRefresh:
		m.refreshList()

	case tea.KeyMsg:
		switch m.state {
		case StateList:
			switch {
			case key.Matches(typedMsg, m.keys.New):
				m.showCreateForm()
				return m, nil
			case key.Matches(typedMsg, m.keys.Edit):
				if len(m.rules) > 0 && m.list.Index() < len(m.rules) {
					m.selected = m.rules[m.list.Index()]
					m.showEditForm()
				}
				return m, nil
			case key.Matches(typedMsg, m.keys.Delete):
				if len(m.rules) > 0 && m.list.Index() < len(m.rules) {
					m.selected = m.rules[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
			}

		case StateForm:
			switch {
			case key.Matches(typedMsg, m.keys.Enter):
				// 检查是否有展开的下拉框或当前聚焦的是TextArea
				if m.form != nil && (m.form.HasExpandedSelect() || m.form.IsTextAreaFocused()) {
					break
				}
				// 焦点不在按钮上时，不处理 Enter 键（让 Form.Update 处理导航）
				if m.form != nil && !m.form.focusOnButtons {
					break
				}
				// 检查是否点击取消按钮
				if m.form != nil && m.form.IsCancelled() {
					m.state = StateList
					m.form = nil
					m.selected = nil
					return m, nil
				}
				// 焦点在确认按钮上才保存
				if m.form != nil && m.form.IsConfirmed() {
					return m, m.saveRule()
				}
				return m, nil
			case key.Matches(typedMsg, m.keys.Esc):
				m.state = StateList
				m.form = nil
				m.selected = nil
				return m, nil
			}

		case StateConfirm:
			switch {
			case key.Matches(typedMsg, m.keys.Enter):
				m.state = StateList
				return m, m.deleteRule()
			case key.Matches(typedMsg, m.keys.Esc):
				m.state = StateList
				m.selected = nil
				return m, nil
			}
		}
	}

	// 更新列表或表单
	switch m.state {
	case StateList:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case StateForm:
		if m.form != nil {
			var cmd tea.Cmd
			m.form, cmd = m.form.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View 渲染模型
func (m *RouteRulesModel) View() string {
	// 构建上下文标题
	contextTitle := ""
	if m.contextSourceName != "" && m.contextProjName != "" {
		contextTitle = fmt.Sprintf(" [%s/%s]", m.contextSourceName, m.contextProjName)
	}

	switch m.state {
	case StateList:
		// 如果有上下文，显示上下文信息
		if contextTitle != "" {
			header := Header(fmt.Sprintf("路由规则列表%s", contextTitle))
			if len(m.rules) == 0 {
				return header + "\n\n" + EmptyListMessage("暂无路由规则，按 'n' 创建新规则") + "\n\n" +
					Help("n 新建", "e 编辑", "d 删除", "Esc 返回项目列表")
			}
			items := m.list.Items()
			return header + "\n\n" + RenderSimpleList(items, m.list.Index(), m.height-8) +
				"\n" + Help("n 新建", "e 编辑", "d 删除", "Esc 返回项目列表")
		}

		if len(m.rules) == 0 {
			return EmptyListMessage("暂无路由规则，按 'n' 创建新规则") + "\n\n" +
				Help("n 新建", "e 编辑", "d 删除")
		}
		items := m.list.Items()
		return RenderSimpleList(items, m.list.Index(), m.height-5) +
			"\n" + Help("n 新建", "e 编辑", "d 删除")

	case StateForm:
		if m.form != nil {
			return m.form.View()
		}

	case StateConfirm:
		if m.selected != nil {
			return Box("确认删除",
				fmt.Sprintf("确定要删除路由规则 '%s' 吗？", m.selected.Name),
				true,
			) + "\n\n" + Help("Enter 确认删除", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm 显示创建表单
func (m *RouteRulesModel) showCreateForm() {
	// 构建匹配类型选项
	matchTypes := []string{MatchTypeParamPath, MatchTypeURLParam, MatchTypeRequestBody, MatchTypeRequestForm, MatchTypePlugin}
	matchLabels := make([]string, len(matchTypes))
	for i, t := range matchTypes {
		matchLabels[i] = matchTypeLabels[t]
	}

	// 构建插件选项
	pluginOptions := make([]string, len(m.plugins)+1)
	pluginLabels := make([]string, len(m.plugins)+1)
	pluginOptions[0] = ""
	pluginLabels[0] = "无"
	for i, p := range m.plugins {
		pluginOptions[i+1] = p.Name
		pluginLabels[i+1] = fmt.Sprintf("%s (%s)", p.Name, p.Version.String)
	}

	// 判断是否有筛选的项目
	hasSelectedProject := m.selectedProj != nil

	// 表单标题
	formTitle := "创建路由规则"
	if hasSelectedProject {
		formTitle = fmt.Sprintf("创建路由规则 [%s/%s]", m.contextSourceName, m.contextProjName)
	}

	fields := []FormField{}

	// 如果没有筛选的项目，显示项目选择字段
	if !hasSelectedProject {
		// 构建项目选项（显示格式：来源/项目）
		projectOptions := make([]string, len(m.projects))
		projectLabels := make([]string, len(m.projects))
		for i, p := range m.projects {
			projectOptions[i] = p.ID
			// 查找项目所属来源名称
			sourceName := "未知来源"
			for _, s := range m.sources {
				if s.ID == p.SourceID {
					sourceName = s.Name
					break
				}
			}
			projectLabels[i] = fmt.Sprintf("%s/%s", sourceName, p.Name)
		}
		fields = append(fields, NewIDSelectField("项目", "project_id", projectOptions, projectLabels, 0))
	}

	fields = append(fields,
		&InputField{
			Label:    "规则名称",
			Key:      "name",
			Input:    newTextInput("例如: weixin_callback"),
			Required: true,
		},
		NewIDChoiceField("匹配类型", "match_type", matchTypes, matchLabels, 0),
		&InputField{
			Label:    "路径模式(完整路径)",
			Key:      "path_pattern",
			Input:    newTextInput("例如: /weixin/callback/{id}"),
			Required: false,
		},
		NewTextAreaField("CEL表达式", "cel_expression", "例如: request.body.event == 'message'", false),
		NewIDSelectField("插件", "plugin_name", pluginOptions, pluginLabels, 0),
		&InputField{
			Label:    "目标URL",
			Key:      "target_url",
			Input:    newTextInput("例如: http://example.com/api"),
			Required: true,
		},
		&NumberField{
			Label: "优先级",
			Key:   "priority",
			Input: newTextInput("0"),
		},
	)

	m.form = NewFormWithFields(formTitle, fields)

	// 设置字段可见性规则
	// path_pattern 仅在 param_path 类型时显示
	m.form.SetVisibilityRule("path_pattern", func(f *Form) bool {
		return f.GetFieldValue("match_type") == MatchTypeParamPath
	})

	// cel_expression 仅在 plugin 类型时不显示（plugin使用自己的匹配逻辑）
	m.form.SetVisibilityRule("cel_expression", func(f *Form) bool {
		return f.GetFieldValue("match_type") != MatchTypePlugin
	})

	// plugin_name 仅在 plugin 类型时显示
	m.form.SetVisibilityRule("plugin_name", func(f *Form) bool {
		return f.GetFieldValue("match_type") == MatchTypePlugin
	})

	// 设置表单尺寸
	if m.height > 0 {
		m.form.SetSize(m.width, m.height)
	}

	m.state = StateForm
}

// showEditForm 显示编辑表单
func (m *RouteRulesModel) showEditForm() {
	if m.selected == nil {
		return
	}

	// 构建项目选项（显示格式：来源/项目）
	projectOptions := make([]string, len(m.projects))
	projectLabels := make([]string, len(m.projects))
	selectedProjectIndex := 0
	for i, p := range m.projects {
		projectOptions[i] = p.ID
		// 查找项目所属来源名称
		sourceName := "未知来源"
		for _, s := range m.sources {
			if s.ID == p.SourceID {
				sourceName = s.Name
				break
			}
		}
		projectLabels[i] = fmt.Sprintf("%s/%s", sourceName, p.Name)
		if p.ID == m.selected.ProjectID {
			selectedProjectIndex = i
		}
	}

	// 构建匹配类型选项
	matchTypes := []string{MatchTypeParamPath, MatchTypeURLParam, MatchTypeRequestBody, MatchTypeRequestForm, MatchTypePlugin}
	matchLabels := make([]string, len(matchTypes))
	for i, t := range matchTypes {
		matchLabels[i] = matchTypeLabels[t]
	}

	// 找到当前匹配类型的索引
	selectedMatchIndex := 0
	for i, t := range matchTypes {
		if t == m.selected.MatchType {
			selectedMatchIndex = i
			break
		}
	}

	// 构建插件选项
	pluginOptions := make([]string, len(m.plugins)+1)
	pluginLabels := make([]string, len(m.plugins)+1)
	pluginOptions[0] = ""
	pluginLabels[0] = "无"
	selectedPluginIndex := 0
	for i, p := range m.plugins {
		pluginOptions[i+1] = p.Name
		pluginLabels[i+1] = fmt.Sprintf("%s (%s)", p.Name, p.Version.String)
		if p.Name == m.selected.PluginName.String {
			selectedPluginIndex = i + 1
		}
	}

	fields := []FormField{
		NewIDSelectField("项目", "project_id", projectOptions, projectLabels, selectedProjectIndex),
		&InputField{
			Label:    "规则名称",
			Key:      "name",
			Input:    newTextInput(""),
			Required: true,
		},
		NewIDChoiceField("匹配类型", "match_type", matchTypes, matchLabels, selectedMatchIndex),
		&InputField{
			Label:    "路径模式(完整路径)",
			Key:      "path_pattern",
			Input:    newTextInput(""),
			Required: false,
		},
		NewTextAreaField("CEL表达式", "cel_expression", "", false),
		NewIDSelectField("插件", "plugin_name", pluginOptions, pluginLabels, selectedPluginIndex),
		&InputField{
			Label:    "目标URL",
			Key:      "target_url",
			Input:    newTextInput(""),
			Required: true,
		},
		&NumberField{
			Label: "优先级",
			Key:   "priority",
			Input: newTextInput(fmt.Sprintf("%d", m.selected.Priority.Int64)),
		},
	}

	m.form = NewFormWithFields("编辑路由规则", fields)

	// 设置字段可见性规则
	m.form.SetVisibilityRule("path_pattern", func(f *Form) bool {
		return f.GetFieldValue("match_type") == MatchTypeParamPath
	})
	m.form.SetVisibilityRule("cel_expression", func(f *Form) bool {
		return f.GetFieldValue("match_type") != MatchTypePlugin
	})
	m.form.SetVisibilityRule("plugin_name", func(f *Form) bool {
		return f.GetFieldValue("match_type") == MatchTypePlugin
	})

	// 设置表单值
	m.form.SetValue("name", m.selected.Name)
	m.form.SetValue("path_pattern", m.selected.PathPattern.String)
	m.form.SetValue("cel_expression", m.selected.CelExpression.String)
	m.form.SetValue("priority", fmt.Sprintf("%d", m.selected.Priority.Int64))

	// 设置表单尺寸
	if m.height > 0 {
		m.form.SetSize(m.width, m.height)
	}

	m.state = StateForm
}

// saveRule 保存路由规则
func (m *RouteRulesModel) saveRule() tea.Cmd {
	return func() tea.Msg {
		if m.form == nil {
			return nil
		}

		if err := m.form.Validate(); err != nil {
			return MsgError{Err: err}
		}

		values := m.form.Values()
		queries := repository.New()

		// 解析优先级
		priority := int64(0)
		fmt.Sscanf(values["priority"], "%d", &priority)

		// 获取项目ID：如果有筛选项目，使用筛选项目的ID；否则使用表单中的项目选择
		projectID := values["project_id"]
		if m.selectedProj != nil {
			projectID = m.selectedProj.ID
		}

		if m.selected != nil {
			// 更新现有规则
			params := &db.UpdateRouteRuleParams{
				Name:          values["name"],
				MatchType:     values["match_type"],
				PathPattern:   sql.NullString{String: values["path_pattern"], Valid: values["path_pattern"] != ""},
				CelExpression: sql.NullString{String: values["cel_expression"], Valid: values["cel_expression"] != ""},
				PluginName:    sql.NullString{String: values["plugin_name"], Valid: values["plugin_name"] != ""},
				Priority:      sql.NullInt64{Int64: priority, Valid: true},
				ID:            m.selected.ID,
			}
			_, err := queries.UpdateRouteRule(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
			m.selected = nil
		} else {
			// 创建新规则
			params := &db.CreateRouteRuleParams{
				ID:            generateID(),
				ProjectID:     projectID,
				Name:          values["name"],
				MatchType:     values["match_type"],
				PathPattern:   sql.NullString{String: values["path_pattern"], Valid: values["path_pattern"] != ""},
				CelExpression: sql.NullString{String: values["cel_expression"], Valid: values["cel_expression"] != ""},
				PluginName:    sql.NullString{String: values["plugin_name"], Valid: values["plugin_name"] != ""},
				Priority:      sql.NullInt64{Int64: priority, Valid: true},
			}
			_, err := queries.CreateRouteRule(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
		}

		m.form = nil
		m.state = StateList
		return tea.Batch(m.loadRules(), SendSuccess("路由规则已保存"))()
	}
}

// deleteRule 删除选中的路由规则
func (m *RouteRulesModel) deleteRule() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteRouteRule(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadRules(), SendSuccess("路由规则已删除"))()
	}
}

// SetFilterProject 设置筛选的项目
func (m *RouteRulesModel) SetFilterProject(project *db.Project) {
	m.selectedProj = project
}

// GetSelectedRule 获取当前选中的路由规则
func (m *RouteRulesModel) GetSelectedRule() *db.RouteRule {
	if len(m.rules) > 0 && m.list.Index() < len(m.rules) {
		return m.rules[m.list.Index()]
	}
	return nil
}

// GetState 获取当前状态
func (m *RouteRulesModel) GetState() AppState {
	return m.state
}

// SetContextInfo 设置上下文信息（来源和项目名称）
func (m *RouteRulesModel) SetContextInfo(sourceName, projectName string) {
	m.contextSourceName = sourceName
	m.contextProjName = projectName
}

// ClearContext 清除上下文信息
func (m *RouteRulesModel) ClearContext() {
	m.selectedProj = nil
	m.contextSourceName = ""
	m.contextProjName = ""
}
