package tui

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/repository"
)

// TabMode 表示当前的 tab
type TabMode int

const (
	TabSources TabMode = iota // 来源 tab
	TabProjects               // 项目 tab
	TabRouteRules             // 路由规则 tab
	TabGlobalConfig           // 全局配置 tab
)

// tab 名称
var tabNames = []string{"来源", "项目", "路由规则", "系统配置"}

// AppState 表示应用状态
type AppState int

const (
	StateList AppState = iota
	StateForm
	StateConfirm
	StateError
)

// App 是主TUI应用
type App struct {
	width  int
	height int
	keys   KeyMap
	state  AppState

	// 当前 tab
	currentTab TabMode

	// 三层数据
	sources  []*db.Source
	projects []*db.Project
	rules    []*db.RouteRule
	plugins  []*db.PluginRegistry

	// 选择状态
	selectedSourceIndex  int
	selectedProjectIndex int
	selectedRuleIndex    int

	// 表单相关
	form        *Form
	formType    string      // "source", "project", "rule"
	editingItem interface{} // 正在编辑的项

	// 删除确认
	deleteItem interface{}

	// 全局配置模型
	globalConfigModel *GlobalConfigModel

	// 路由规则模型
	routeRulesModel *RouteRulesModel

	// 消息
	err     error
	success string
}

// NewApp 创建新的TUI应用
func NewApp() *App {
	return &App{
		keys:                 DefaultKeyMap(),
		state:                StateList,
		currentTab:           TabSources,
		selectedSourceIndex:  0,
		selectedProjectIndex: 0,
		selectedRuleIndex:    0,
		globalConfigModel:    NewGlobalConfigModel(),
		routeRulesModel:      NewRouteRulesModel(),
	}
}

// Init 初始化应用
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadSources(),
		a.loadPlugins(),
		a.globalConfigModel.Init(),
		a.routeRulesModel.Init(),
	)
}

// loadSources 加载来源列表
func (a *App) loadSources() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		sources, err := queries.ListSources(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgSourcesLoaded{Sources: sources}
	}
}

// loadProjects 加载项目列表
func (a *App) loadProjects(sourceID string) tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		projects, err := queries.ListProjectsBySourceID(context.Background(), sourceID)
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgProjectsLoaded{Projects: projects}
	}
}

// loadRules 加载路由规则
func (a *App) loadRules(projectID string) tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		rules, err := queries.ListRouteRulesByProjectID(context.Background(), projectID)
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgRouteRulesLoaded{Rules: rules}
	}
}

// loadProjectRules 加载项目的路由规则（用于项目编辑界面）
func (a *App) loadProjectRules(projectID string) tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		rules, err := queries.ListRouteRulesByProjectID(context.Background(), projectID)
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgProjectRulesLoaded{Rules: rules}
	}
}

// loadPlugins 加载插件列表
func (a *App) loadPlugins() tea.Cmd {
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

// Update 处理消息
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.globalConfigModel.SetSize(msg.Width, msg.Height)
		a.routeRulesModel.SetSize(msg.Width, msg.Height)

	case MsgSourcesLoaded:
		a.sources = msg.Sources
		// 如果有来源，加载第一个来源的项目
		if len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
			cmds = append(cmds, a.loadProjects(a.sources[a.selectedSourceIndex].ID))
		}

	case MsgProjectsLoaded:
		a.projects = msg.Projects
		// 如果有项目，加载第一个项目的规则
		if len(a.projects) > 0 && a.selectedProjectIndex < len(a.projects) {
			cmds = append(cmds, a.loadRules(a.projects[a.selectedProjectIndex].ID))
		} else {
			a.rules = nil
		}

	case MsgPluginsLoaded:
		a.plugins = msg.Plugins

	case MsgRouteRulesLoaded:
		a.rules = msg.Rules

	case MsgError:
		a.err = msg.Err

	case MsgSuccess:
		a.success = msg.Message
		a.err = nil
		a.state = StateList

	case MsgManageRules:
		// 从项目列表进入该项目的路由规则管理
		a.currentTab = TabRouteRules
		a.routeRulesModel.SetFilterProject(msg.Project)
		a.routeRulesModel.SetContextInfo(msg.SourceName, msg.Project.Name)
		// 加载该项目的路由规则
		cmds = append(cmds, a.loadRules(msg.Project.ID))

	case tea.KeyMsg:
		if a.state == StateForm {
			return a.handleFormKeys(msg)
		} else if a.state == StateConfirm {
			return a.handleConfirmKeys(msg)
		}
		return a.handleListKeys(msg)
	}

	return a, tea.Batch(cmds...)
}

// handleListKeys 处理列表模式下的按键
func (a *App) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 如果在全局配置 tab，将按键传递给 globalConfigModel 处理
	if a.currentTab == TabGlobalConfig {
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Tab):
			return a, a.switchTab(1)
		case key.Matches(msg, a.keys.Left):
			return a, a.switchTab(-1)
		case key.Matches(msg, a.keys.Right):
			return a, a.switchTab(1)
		default:
			// 其他按键交给 globalConfigModel 处理
			model, cmd := a.globalConfigModel.Update(msg)
			if m, ok := model.(*GlobalConfigModel); ok {
				a.globalConfigModel = m
			}
			return a, cmd
		}
	}

	// 如果在路由规则 tab，将按键传递给 routeRulesModel 处理
	if a.currentTab == TabRouteRules {
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Tab):
			return a, a.switchTab(1)
		case key.Matches(msg, a.keys.Left):
			return a, a.switchTab(-1)
		case key.Matches(msg, a.keys.Right):
			return a, a.switchTab(1)
		default:
			// 其他按键交给 routeRulesModel 处理
			model, cmd := a.routeRulesModel.Update(msg)
			if m, ok := model.(*RouteRulesModel); ok {
				a.routeRulesModel = m
			}
			return a, cmd
		}
	}

	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.Tab):
		return a, a.switchTab(1)

	case key.Matches(msg, a.keys.Left):
		return a, a.switchTab(-1)

	case key.Matches(msg, a.keys.Right):
		return a, a.switchTab(1)

	case key.Matches(msg, a.keys.Down):
		a.moveDown()

	case key.Matches(msg, a.keys.Up):
		a.moveUp()

	case key.Matches(msg, a.keys.Enter):
		return a, a.handleEnter()

	case key.Matches(msg, a.keys.Esc):
		// 在 tab 模式下，Esc 不做任何操作

	case key.Matches(msg, a.keys.New):
		return a, a.showCreateForm()

	case key.Matches(msg, a.keys.Edit):
		return a, a.showEditForm()

	case key.Matches(msg, a.keys.Delete):
		return a, a.showDeleteConfirm()
	}

	return a, nil
}

// switchTab 切换 tab，direction: 1 为下一个，-1 为上一个
func (a *App) switchTab(direction int) tea.Cmd {
	newTab := int(a.currentTab) + direction
	if newTab < 0 {
		newTab = 3
	} else if newTab > 3 {
		newTab = 0
	}
	a.currentTab = TabMode(newTab)
	// 切换到全局配置 tab 时刷新配置
	if a.currentTab == TabGlobalConfig {
		a.globalConfigModel.Update(MsgRefresh{})
	}
	// 切换到路由规则 tab 时刷新规则列表
	if a.currentTab == TabRouteRules {
		a.routeRulesModel.Update(MsgRefresh{})
	}
	// 切换到项目 tab 时，加载当前选中来源的项目
	if a.currentTab == TabProjects && len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
		return a.loadProjects(a.sources[a.selectedSourceIndex].ID)
	}
	return nil
}

// handleEnter 处理 Enter 键
func (a *App) handleEnter() tea.Cmd {
	switch a.currentTab {
	case TabSources:
		// 在来源 tab，Enter 键加载该来源的项目
		if len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
			return a.loadProjects(a.sources[a.selectedSourceIndex].ID)
		}
	case TabProjects:
		// 在项目 tab，Enter 键加载该项目的规则
		if len(a.projects) > 0 && a.selectedProjectIndex < len(a.projects) {
			return a.loadRules(a.projects[a.selectedProjectIndex].ID)
		}
	}
	return nil
}

// handleFormKeys 处理表单模式下的按键
func (a *App) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit

	case key.Matches(msg, a.keys.Esc):
		a.state = StateList
		a.form = nil
		a.formType = ""
		a.editingItem = nil
		return a, nil

	case key.Matches(msg, a.keys.Enter):
		if a.form != nil && (a.form.HasExpandedSelect() || a.form.IsTextAreaFocused()) {
			var cmd tea.Cmd
			a.form, cmd = a.form.Update(msg)
			return a, cmd
		}
		// 焦点不在按钮上时，不处理 Enter 键（让 Form.Update 处理导航）
		if a.form != nil && !a.form.focusOnButtons {
			var cmd tea.Cmd
			a.form, cmd = a.form.Update(msg)
			return a, cmd
		}
		// 检查是否取消了表单（焦点在取消按钮上）
		if a.form != nil && a.form.IsCancelled() {
			a.state = StateList
			a.form = nil
			a.formType = ""
			a.editingItem = nil
			return a, nil
		}
		// 焦点在确认按钮上，保存表单
		if a.form != nil && a.form.IsConfirmed() {
			return a, a.saveForm()
		}
		return a, nil
	}

	if a.form != nil {
		var cmd tea.Cmd
		a.form, cmd = a.form.Update(msg)
		return a, cmd
	}

	return a, nil
}

// handleConfirmKeys 处理确认模式下的按键
func (a *App) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Enter):
		a.state = StateList
		return a, a.deleteItemAction()

	case key.Matches(msg, a.keys.Esc):
		a.state = StateList
		a.deleteItem = nil
		return a, nil
	}

	return a, nil
}

// moveDown 向下移动
func (a *App) moveDown() {
	switch a.currentTab {
	case TabSources:
		if a.selectedSourceIndex < len(a.sources)-1 {
			a.selectedSourceIndex++
		}
	case TabProjects:
		if a.selectedProjectIndex < len(a.projects)-1 {
			a.selectedProjectIndex++
		}
	case TabRouteRules:
		// 路由规则 tab 由 routeRulesModel 处理
	case TabGlobalConfig:
		// 全局配置 tab 暂无列表
	}
}

// moveUp 向上移动
func (a *App) moveUp() {
	switch a.currentTab {
	case TabSources:
		if a.selectedSourceIndex > 0 {
			a.selectedSourceIndex--
		}
	case TabProjects:
		if a.selectedProjectIndex > 0 {
			a.selectedProjectIndex--
		}
	case TabRouteRules:
		// 路由规则 tab 由 routeRulesModel 处理
	case TabGlobalConfig:
		// 全局配置 tab 暂无列表
	}
}

// showCreateForm 显示创建表单
func (a *App) showCreateForm() tea.Cmd {
	switch a.currentTab {
	case TabSources:
		a.form = NewFormWithFields("创建来源", []FormField{
			&InputField{
				Label:    "名称",
				Key:      "name",
				Input:    newTextInput("例如: weixin"),
				Required: true,
			},
			&InputField{
				Label:    "描述",
				Key:      "description",
				Input:    newTextInput(""),
				Required: false,
			},
		})
		a.formType = "source"
		a.state = StateForm
		return nil

	case TabProjects:
		if len(a.sources) == 0 {
			a.err = fmt.Errorf("请先创建来源")
			return nil
		}

		// 构建来源下拉选项
		sourceOptions := make([]string, len(a.sources))
		sourceLabels := make([]string, len(a.sources))
		for i, s := range a.sources {
			sourceOptions[i] = s.ID
			sourceLabels[i] = s.Name
		}

		a.form = NewFormWithFields("创建项目", []FormField{
			NewIDSelectField("来源", "source_id", sourceOptions, sourceLabels, a.selectedSourceIndex),
			&InputField{
				Label:    "名称",
				Key:      "name",
				Input:    newTextInput("例如: callback"),
				Required: true,
			},
			&InputField{
				Label:    "描述",
				Key:      "description",
				Input:    newTextInput(""),
				Required: false,
			},
			&InputField{
				Label:    "目标URL",
				Key:      "target_url",
				Input:    newTextInput("例如: http://backend/api"),
				Required: false,
			},
			&NumberField{
				Label: "优先级",
				Key:   "priority",
				Input: newTextInput("0"),
			},
		})
		a.formType = "project"
		a.state = StateForm
		return nil

	case TabRouteRules:
		// 路由规则 tab 由 routeRulesModel 处理
		a.err = fmt.Errorf("请使用路由规则 tab 的表单")
		return nil

	case TabGlobalConfig:
		// 全局配置 tab 暂不支持创建
		a.err = fmt.Errorf("全局配置不支持创建操作")
		return nil
	}

	return nil
}

// showEditForm 显示编辑表单
func (a *App) showEditForm() tea.Cmd {
	switch a.currentTab {
	case TabSources:
		if len(a.sources) == 0 || a.selectedSourceIndex >= len(a.sources) {
			return nil
		}
		source := a.sources[a.selectedSourceIndex]
		a.form = NewFormWithFields("编辑来源", []FormField{
			&InputField{
				Label:    "名称",
				Key:      "name",
				Input:    newTextInput(""),
				Required: true,
			},
			&InputField{
				Label:    "描述",
				Key:      "description",
				Input:    newTextInput(""),
				Required: false,
			},
		})
		a.form.SetValue("name", source.Name)
		a.form.SetValue("description", source.Description.String)
		a.formType = "source"
		a.editingItem = source
		a.state = StateForm
		return nil

	case TabProjects:
		if len(a.projects) == 0 || a.selectedProjectIndex >= len(a.projects) {
			return nil
		}
		// 确保有来源数据
		if len(a.sources) == 0 {
			a.err = fmt.Errorf("来源数据未加载，请稍后重试")
			return nil
		}
		project := a.projects[a.selectedProjectIndex]

		// 构建来源下拉选项
		sourceOptions := make([]string, len(a.sources))
		sourceLabels := make([]string, len(a.sources))
		currentSourceIndex := 0
		for i, s := range a.sources {
			sourceOptions[i] = s.ID
			sourceLabels[i] = s.Name
			if s.ID == project.SourceID {
				currentSourceIndex = i
			}
		}

		a.form = NewFormWithFields("编辑项目", []FormField{
			NewIDSelectField("来源", "source_id", sourceOptions, sourceLabels, currentSourceIndex),
			&InputField{
				Label:    "名称",
				Key:      "name",
				Input:    newTextInput(""),
				Required: true,
			},
			&InputField{
				Label:    "描述",
				Key:      "description",
				Input:    newTextInput(""),
				Required: false,
			},
			&InputField{
				Label:    "目标URL",
				Key:      "target_url",
				Input:    newTextInput(""),
				Required: false,
			},
			&NumberField{
				Label: "优先级",
				Key:   "priority",
				Input: newTextInput(fmt.Sprintf("%d", project.Priority.Int64)),
			},
		})
		a.form.SetValue("name", project.Name)
		a.form.SetValue("description", project.Description.String)
		a.form.SetValue("target_url", project.TargetUrl.String)
		a.formType = "project"
		a.editingItem = project
		a.state = StateForm
		return nil

	case TabRouteRules:
		// 路由规则 tab 由 routeRulesModel 处理
		a.err = fmt.Errorf("请使用路由规则 tab 的表单")
		return nil

	case TabGlobalConfig:
		// 全局配置 tab 暂不支持编辑
		a.err = fmt.Errorf("全局配置不支持编辑操作")
		return nil
	}

	return nil
}

// showDeleteConfirm 显示删除确认
func (a *App) showDeleteConfirm() tea.Cmd {
	switch a.currentTab {
	case TabSources:
		if len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
			a.deleteItem = a.sources[a.selectedSourceIndex]
			a.state = StateConfirm
		}
	case TabProjects:
		if len(a.projects) > 0 && a.selectedProjectIndex < len(a.projects) {
			a.deleteItem = a.projects[a.selectedProjectIndex]
			a.state = StateConfirm
		}
	case TabRouteRules:
		// 路由规则 tab 由 routeRulesModel 处理
		a.err = fmt.Errorf("请使用路由规则 tab 的删除功能")
	case TabGlobalConfig:
		// 全局配置 tab 暂不支持删除
		a.err = fmt.Errorf("全局配置不支持删除操作")
	}
	return nil
}

// saveForm 保存表单
func (a *App) saveForm() tea.Cmd {
	return func() tea.Msg {
		if a.form == nil {
			return nil
		}

		if err := a.form.Validate(); err != nil {
			return MsgError{Err: err}
		}

		values := a.form.Values()
		queries := repository.New()

		switch a.formType {
		case "source":
			return a.saveSource(queries, values)
		case "project":
			return a.saveProject(queries, values)
		case "rule":
			return a.saveRule(queries, values)
		}

		return nil
	}
}

// saveSource 保存来源
func (a *App) saveSource(queries *db.Queries, values map[string]string) tea.Msg {
	if source, ok := a.editingItem.(*db.Source); ok {
		// 更新
		params := &db.UpdateSourceParams{
			Name:        values["name"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
			ID:          source.ID,
		}
		_, err := queries.UpdateSource(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}
	} else {
		// 创建
		params := &db.CreateSourceParams{
			ID:          generateID(),
			Name:        values["name"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
		}
		_, err := queries.CreateSource(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}
	}

	a.form = nil
	a.formType = ""
	a.editingItem = nil
	return tea.Batch(a.loadSources(), SendSuccess("来源已保存"))()
}

// saveProject 保存项目
func (a *App) saveProject(queries *db.Queries, values map[string]string) tea.Msg {
	// 从表单获取来源ID，如果没有则使用当前选中的来源
	sourceID := values["source_id"]
	if sourceID == "" && len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
		sourceID = a.sources[a.selectedSourceIndex].ID
	}

	priority := parseInt(values["priority"])
	targetURL := values["target_url"]

	if project, ok := a.editingItem.(*db.Project); ok {
		// 更新
		params := &db.UpdateProjectParams{
			SourceID:    sourceID,
			Name:        values["name"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
			TargetUrl:   sql.NullString{String: targetURL, Valid: targetURL != ""},
			Priority:    sql.NullInt64{Int64: priority, Valid: true},
			ID:          project.ID,
		}
		_, err := queries.UpdateProject(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}
	} else {
		// 创建
		params := &db.CreateProjectParams{
			ID:          generateID(),
			SourceID:    sourceID,
			Name:        values["name"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
			TargetUrl:   sql.NullString{String: targetURL, Valid: targetURL != ""},
			Priority:    sql.NullInt64{Int64: priority, Valid: true},
		}
		_, err := queries.CreateProject(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}
	}

	a.form = nil
	a.formType = ""
	a.editingItem = nil
	if len(a.sources) > 0 {
		return tea.Batch(a.loadProjects(a.sources[a.selectedSourceIndex].ID), SendSuccess("项目已保存"))()
	}
	return SendSuccess("项目已保存")
}

// saveRule 保存路由规则
func (a *App) saveRule(queries *db.Queries, values map[string]string) tea.Msg {
	projectID := ""
	if project, ok := a.editingItem.(*db.Project); ok {
		// 如果是从项目编辑界面保存规则，使用当前项目的ID
		projectID = project.ID
	} else if len(a.projects) > 0 && a.selectedProjectIndex < len(a.projects) {
		projectID = a.projects[a.selectedProjectIndex].ID
	}

	priority := parseInt(values["priority"])

	if rule, ok := a.editingItem.(*db.RouteRule); ok {
		// 更新
		params := &db.UpdateRouteRuleParams{
			Name:          values["name"],
			MatchType:     values["match_type"],
			PathPattern:   sql.NullString{String: values["path_pattern"], Valid: values["path_pattern"] != ""},
			CelExpression: sql.NullString{String: values["cel_expression"], Valid: values["cel_expression"] != ""},
			PluginName:    sql.NullString{String: values["plugin_name"], Valid: values["plugin_name"] != ""},
			Priority:      sql.NullInt64{Int64: priority, Valid: true},
			ID:            rule.ID,
		}
		_, err := queries.UpdateRouteRule(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}
	} else {
		// 创建
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

	a.form = nil
	a.formType = ""
	a.editingItem = nil
	if len(a.projects) > 0 {
		return tea.Batch(a.loadRules(a.projects[a.selectedProjectIndex].ID), SendSuccess("路由规则已保存"))()
	}
	return SendSuccess("路由规则已保存")
}

// deleteItemAction 执行删除
func (a *App) deleteItemAction() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		var err error

		switch item := a.deleteItem.(type) {
		case *db.Source:
			err = queries.DeleteSource(context.Background(), item.ID)
			a.selectedSourceIndex = 0
		case *db.Project:
			err = queries.DeleteProject(context.Background(), item.ID)
			a.selectedProjectIndex = 0
		case *db.RouteRule:
			err = queries.DeleteRouteRule(context.Background(), item.ID)
			a.selectedRuleIndex = 0
			// 如果是从项目编辑界面删除规则，需要刷新 projectRules
			if a.state == StateList && a.formType == "project" {
				// 将在返回后重新加载
			}
		}

		a.deleteItem = nil

		if err != nil {
			return MsgError{Err: err}
		}

		// 重新加载数据
		switch a.currentTab {
		case TabSources:
			return tea.Batch(a.loadSources(), SendSuccess("已删除"))()
		case TabProjects:
			if len(a.sources) > 0 {
				return tea.Batch(a.loadProjects(a.sources[a.selectedSourceIndex].ID), SendSuccess("已删除"))()
			}
		case TabGlobalConfig:
			// 全局配置 tab 暂无删除操作
		}

		return SendSuccess("已删除")
	}
}

// View 渲染应用
func (a *App) View() string {
	if a.state == StateForm && a.form != nil {
		var b strings.Builder
		// 渲染 TabBar
		b.WriteString(TabBar(tabNames, int(a.currentTab)))
		b.WriteString("\n\n")
		// 渲染表单
		b.WriteString(a.form.View())
		return b.String()
	}

	if a.state == StateConfirm && a.deleteItem != nil {
		var b strings.Builder
		// 渲染 TabBar
		b.WriteString(TabBar(tabNames, int(a.currentTab)))
		b.WriteString("\n\n")
		// 渲染确认框
		var name string
		var itemType string
		switch item := a.deleteItem.(type) {
		case *db.Source:
			name = item.Name
			itemType = "来源"
		case *db.Project:
			name = item.Name
			itemType = "项目"
		case *db.RouteRule:
			name = item.Name
			itemType = "路由规则"
		}
		b.WriteString(Box("确认删除",
			fmt.Sprintf("确定要删除%s '%s' 吗？", itemType, name),
			true,
		))
		b.WriteString("\n\n")
		b.WriteString(Help("Enter 确认删除", "Esc 取消"))
		return b.String()
	}

	return a.renderTabView()
}

// renderTabView 渲染 tab 视图
func (a *App) renderTabView() string {
	var b strings.Builder

	// 渲染 TabBar
	b.WriteString(TabBar(tabNames, int(a.currentTab)))
	b.WriteString("\n\n")

	// 根据当前 tab 渲染内容
	switch a.currentTab {
	case TabSources:
		a.renderSourceList(&b)
	case TabProjects:
		a.renderProjectList(&b)
	case TabRouteRules:
		a.renderRouteRules(&b)
	case TabGlobalConfig:
		a.renderGlobalConfig(&b)
	}

	// 错误/成功消息
	if a.err != nil {
		b.WriteString("\n")
		b.WriteString(Error(a.err.Error()))
	}
	if a.success != "" {
		b.WriteString("\n")
		b.WriteString(Success(a.success))
	}

	// 帮助提示
	b.WriteString("\n")
	b.WriteString(a.renderHelp())

	return b.String()
}

// renderSourceList 渲染来源列表
func (a *App) renderSourceList(b *strings.Builder) {
	title := "来源列表"
	b.WriteString(styleSectionTitle.Render(title))
	b.WriteString("\n\n")

	if len(a.sources) == 0 {
		b.WriteString(styleEmptyState.Render("  暂无来源，按 'n' 创建"))
	} else {
		for i, source := range a.sources {
			selected := i == a.selectedSourceIndex
			prefix := "  "
			if selected {
				prefix = "▸ "
			}
			line := prefix + source.Name
			if source.Description.Valid && source.Description.String != "" {
				line += " - " + Truncate(source.Description.String, 40)
			}
			if selected {
				b.WriteString(styleItemSelected.Render(line))
			} else {
				b.WriteString(styleItem.Render(line))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
}

// renderProjectList 渲染项目列表
func (a *App) renderProjectList(b *strings.Builder) {
	// 标题显示当前来源
	sourceName := ""
	if len(a.sources) > 0 && a.selectedSourceIndex < len(a.sources) {
		sourceName = a.sources[a.selectedSourceIndex].Name
	}

	title := "项目列表"
	if sourceName != "" {
		title = fmt.Sprintf("项目列表 (来源: %s)", sourceName)
	}
	b.WriteString(styleSectionTitle.Render(title))
	b.WriteString("\n\n")

	if len(a.projects) == 0 {
		if sourceName != "" {
			b.WriteString(styleEmptyState.Render("  暂无项目，按 'n' 创建"))
		} else {
			b.WriteString(styleEmptyState.Render("  请先在「来源」tab 中选择来源"))
		}
	} else {
		for i, project := range a.projects {
			selected := i == a.selectedProjectIndex
			prefix := "  "
			if selected {
				prefix = "▸ "
			}
			line := prefix + project.Name
			if project.Description.Valid && project.Description.String != "" {
				line += " - " + Truncate(project.Description.String, 30)
			}
			if selected {
				b.WriteString(styleItemSelected.Render(line))
			} else {
				b.WriteString(styleItem.Render(line))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
}

// renderGlobalConfig 渲染全局配置
func (a *App) renderGlobalConfig(b *strings.Builder) {
	b.WriteString(a.globalConfigModel.View())
	b.WriteString("\n")
}

// renderRouteRules 渲染路由规则列表
func (a *App) renderRouteRules(b *strings.Builder) {
	b.WriteString(a.routeRulesModel.View())
	b.WriteString("\n")
}

// renderHelp 渲染帮助提示
func (a *App) renderHelp() string {
	baseHelp := Help("←→ 切换tab", "↑↓ 导航", "n 新建", "e 编辑", "d 删除", "q 退出")
	switch a.currentTab {
	case TabSources:
		return baseHelp
	case TabProjects:
		return Help("←→ 切换tab", "↑↓ 导航", "Enter 加载规则", "n 新建", "e 编辑", "d 删除", "q 退出")
	case TabRouteRules:
		return Help("←→ 切换tab", "↑↓ 导航", "n 新建", "e 编辑", "d 删除", "Esc 返回", "q 退出")
	case TabGlobalConfig:
		return Help("←→ 切换tab", "↑↓ 导航", "Space/Enter 切换开关", "q 退出")
	}
	return baseHelp
}

// parseInt 解析整数
func parseInt(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

// Run 启动TUI应用
func Run() error {
	app := NewApp()
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}