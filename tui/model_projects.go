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

// ProjectsModel 管理项目列表
type ProjectsModel struct {
	list     list.Model
	projects []*db.Project
	sources  []*db.Source // 来源列表，用于选择器
	state    AppState
	form     *Form
	selected *db.Project
	width    int
	height   int
	keys     KeyMap
}

// NewProjectsModel 创建一个新的项目模型
func NewProjectsModel() *ProjectsModel {
	m := &ProjectsModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "项目列表", 80, 20)
	return m
}

// Init 初始化模型
func (m *ProjectsModel) Init() tea.Cmd {
	return tea.Batch(m.loadSources(), m.loadProjects())
}

// SetSize 设置尺寸
func (m *ProjectsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
	// 如果表单存在，也设置表单尺寸
	if m.form != nil {
		m.form.SetSize(width, height)
	}
}

// MsgSourcesLoadedForProject 项目模块的来源加载完成消息
type MsgSourcesLoadedForProject struct {
	Sources []*db.Source
}

// loadSources 加载来源列表
func (m *ProjectsModel) loadSources() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return nil
		}
		sources, err := queries.ListSources(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgSourcesLoadedForProject{Sources: sources}
	}
}

// MsgProjectsLoaded 项目加载完成消息
type MsgProjectsLoaded struct {
	Projects []*db.Project
}

// loadProjects 从数据库加载项目
func (m *ProjectsModel) loadProjects() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}

		projects, err := queries.ListProjects(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgProjectsLoaded{Projects: projects}
	}
}

// refreshList 刷新列表
func (m *ProjectsModel) refreshList() {
	items := make([]list.Item, len(m.projects))
	for i, p := range m.projects {
		// 获取来源名称
		sourceName := "未知来源"
		for _, s := range m.sources {
			if s.ID == p.SourceID {
				sourceName = s.Name
				break
			}
		}
		desc := p.Description.String
		items[i] = MenuItem{
			title:       fmt.Sprintf("%s/%s", sourceName, p.Name),
			description: desc,
		}
	}
	m.list.SetItems(items)
}

// Update 处理消息
func (m *ProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MsgSourcesLoadedForProject:
		m.sources = msg.Sources
		m.refreshList()

	case MsgProjectsLoaded:
		m.projects = msg.Projects
		m.refreshList()

	case MsgRefresh:
		m.refreshList()

	case tea.KeyMsg:
		switch m.state {
		case StateList:
			switch {
			case key.Matches(msg, m.keys.New):
				m.showCreateForm()
				return m, nil
			case key.Matches(msg, m.keys.Edit):
				if len(m.projects) > 0 && m.list.Index() < len(m.projects) {
					m.selected = m.projects[m.list.Index()]
					m.showEditForm()
				}
				return m, nil
			case key.Matches(msg, m.keys.Delete):
				if len(m.projects) > 0 && m.list.Index() < len(m.projects) {
					m.selected = m.projects[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
			case key.Matches(msg, m.keys.Toggle):
				if len(m.projects) > 0 && m.list.Index() < len(m.projects) {
					return m, m.toggleProject(m.projects[m.list.Index()])
				}
			}

		case StateForm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				// 检查是否有展开的下拉框
				if m.form != nil && m.form.HasExpandedSelect() {
					// 让表单处理 Enter 键
					break
				}
				return m, m.saveProject()
			case key.Matches(msg, m.keys.Esc):
				m.state = StateList
				m.form = nil
				m.selected = nil
				return m, nil
			}

		case StateConfirm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				m.state = StateList
				return m, m.deleteProject()
			case key.Matches(msg, m.keys.Esc):
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
func (m *ProjectsModel) View() string {
	switch m.state {
	case StateList:
		if len(m.projects) == 0 {
			return EmptyListMessage("暂无项目，按 'n' 创建新项目") + "\n\n" +
				Help("n 新建", "e 编辑", "d 删除", "space 切换状态")
		}
		items := m.list.Items()
		return RenderSimpleList(items, m.list.Index(), m.height-5) +
			"\n" + Help("n 新建", "e 编辑", "d 删除", "space 切换状态")

	case StateForm:
		if m.form != nil {
			return m.form.View()
		}

	case StateConfirm:
		if m.selected != nil {
			return Box("确认删除",
				fmt.Sprintf("确定要删除项目 '%s' 吗？\n\n此操作将同时删除关联的路由规则。", m.selected.Name),
				true,
			) + "\n\n" + Help("Enter 确认删除", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm 显示创建表单
func (m *ProjectsModel) showCreateForm() {
	// 构建来源选项
	sourceOptions := make([]string, len(m.sources))
	sourceLabels := make([]string, len(m.sources))
	for i, s := range m.sources {
		sourceOptions[i] = s.ID
		sourceLabels[i] = s.Name
	}

	m.form = NewFormWithFields("创建项目", []FormField{
		NewIDSelectField("来源", "source_id", sourceOptions, sourceLabels, 0),
		&InputField{
			Label:    "名称",
			Key:      "name",
			Input:    newTextInput("例如: callback"),
			Required: true,
		},
		&InputField{
			Label:    "描述",
			Key:      "description",
			Input:    newTextInput("例如: 微信回调项目"),
			Required: false,
		},
	})
	// 设置表单尺寸
	if m.height > 0 {
		m.form.SetSize(m.width, m.height)
	}
	m.state = StateForm
}

// showEditForm 显示编辑表单
func (m *ProjectsModel) showEditForm() {
	if m.selected == nil {
		return
	}

	// 构建来源选项
	sourceOptions := make([]string, len(m.sources))
	sourceLabels := make([]string, len(m.sources))
	selectedIndex := 0
	for i, s := range m.sources {
		sourceOptions[i] = s.ID
		sourceLabels[i] = s.Name
		if s.ID == m.selected.SourceID {
			selectedIndex = i
		}
	}

	m.form = NewFormWithFields("编辑项目", []FormField{
		NewIDSelectField("来源", "source_id", sourceOptions, sourceLabels, selectedIndex),
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
	// 设置表单值
	m.form.SetValue("name", m.selected.Name)
	m.form.SetValue("description", m.selected.Description.String)
	// 设置表单尺寸
	if m.height > 0 {
		m.form.SetSize(m.width, m.height)
	}
	m.state = StateForm
}

// saveProject 保存项目
func (m *ProjectsModel) saveProject() tea.Cmd {
	return func() tea.Msg {
		if m.form == nil {
			return nil
		}

		if err := m.form.Validate(); err != nil {
			return MsgError{Err: err}
		}

		values := m.form.Values()
		queries := repository.New()

		if m.selected != nil {
			// 更新现有项目
			params := &db.UpdateProjectParams{
				Name:        values["name"],
				Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
				Enabled:     m.selected.Enabled,
				ID:          m.selected.ID,
			}
			_, err := queries.UpdateProject(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
			m.selected = nil
		} else {
			// 创建新项目
			params := &db.CreateProjectParams{
				ID:          generateID(),
				SourceID:    values["source_id"],
				Name:        values["name"],
				Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
				Enabled:     sql.NullInt64{Int64: 1, Valid: true},
			}
			_, err := queries.CreateProject(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
		}

		m.form = nil
		m.state = StateList
		return tea.Batch(m.loadProjects(), SendSuccess("项目已保存"))()
	}
}

// deleteProject 删除选中的项目
func (m *ProjectsModel) deleteProject() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteProject(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadProjects(), SendSuccess("项目已删除"))()
	}
}

// toggleProject 切换项目启用状态
func (m *ProjectsModel) toggleProject(project *db.Project) tea.Cmd {
	return func() tea.Msg {
		newEnabled := int64(1)
		if project.Enabled.Int64 == 1 {
			newEnabled = 0
		}

		params := &db.UpdateProjectParams{
			Name:        project.Name,
			Description: project.Description,
			Enabled:     sql.NullInt64{Int64: newEnabled, Valid: true},
			ID:          project.ID,
		}

		queries := repository.New()
		_, err := queries.UpdateProject(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadProjects(), SendSuccess("状态已切换"))()
	}
}

// GetSelectedProject 获取当前选中的项目
func (m *ProjectsModel) GetSelectedProject() *db.Project {
	if len(m.projects) > 0 && m.list.Index() < len(m.projects) {
		return m.projects[m.list.Index()]
	}
	return nil
}

// GetProjects 获取所有项目
func (m *ProjectsModel) GetProjects() []*db.Project {
	return m.projects
}

// GetState 获取当前状态
func (m *ProjectsModel) GetState() AppState {
	return m.state
}
