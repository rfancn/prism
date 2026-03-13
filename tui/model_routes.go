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

// RoutesModel manages the routes list and form
type RoutesModel struct {
	list     list.Model
	routes   []*db.Route
	state    AppState
	form     *Form
	selected *db.Route
	width    int
	height   int
	keys     KeyMap
}

// NewRoutesModel creates a new routes model
func NewRoutesModel() *RoutesModel {
	m := &RoutesModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "路由列表", 80, 20)
	return m
}

// Init initializes the model
func (m *RoutesModel) Init() tea.Cmd {
	return m.loadRoutes()
}

// SetSize sets the size
func (m *RoutesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// loadRoutes loads routes from database
func (m *RoutesModel) loadRoutes() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}
		routes, err := queries.ListRoutes(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		m.routes = routes
		return MsgRefresh{}
	}
}

// refreshList refreshes the list items
func (m *RoutesModel) refreshList() {
	items := make([]list.Item, len(m.routes))
	for i, r := range m.routes {
		status := "启用"
		if r.Enabled.Int64 == 0 {
			status = "禁用"
		}
		items[i] = MenuItem{
			title:       fmt.Sprintf("%s → %s", r.Identifier, Truncate(r.TargetUrl, 30)),
			description: fmt.Sprintf("[%s] %s", status, r.Pattern),
		}
	}
	m.list.SetItems(items)
}

// Update handles messages
func (m *RoutesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
				if len(m.routes) > 0 && m.list.Index() < len(m.routes) {
					m.selected = m.routes[m.list.Index()]
					m.showEditForm()
				}
				return m, nil
			case key.Matches(msg, m.keys.Delete):
				if len(m.routes) > 0 && m.list.Index() < len(m.routes) {
					m.selected = m.routes[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				// View details
				return m, nil
			case key.Matches(msg, m.keys.Toggle):
				// Toggle enabled
				if len(m.routes) > 0 && m.list.Index() < len(m.routes) {
					return m, m.toggleRoute(m.routes[m.list.Index()])
				}
			}

		case StateForm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				return m, m.saveRoute()
			case key.Matches(msg, m.keys.Esc):
				m.state = StateList
				m.form = nil
				return m, nil
			}

		case StateConfirm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				m.state = StateList
				return m, m.deleteRoute()
			case key.Matches(msg, m.keys.Esc):
				m.state = StateList
				return m, nil
			}
		}
	}

	// Update list or form
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

// View renders the model
func (m *RoutesModel) View() string {
	switch m.state {
	case StateList:
		if len(m.routes) == 0 {
			return Header("路由列表") + "\n\n" +
				EmptyListMessage("暂无路由，按 'n' 创建新路由") + "\n\n" +
				Help("n 新建", "e 编辑", "d 删除", "space 切换状态")
		}
		return m.list.View() + "\n\n" +
			Help("n 新建", "e 编辑", "d 删除", "space 切换状态", "enter 详情")

	case StateForm:
		if m.form != nil {
			return m.form.View()
		}

	case StateConfirm:
		if m.selected != nil {
			return Box("确认删除",
				fmt.Sprintf("确定要删除路由 '%s' 吗？\n\n此操作不可撤销。", m.selected.Identifier),
				true,
			) + "\n\n" + Help("Enter 确认删除", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm shows the create form
func (m *RoutesModel) showCreateForm() {
	m.form = NewForm("创建路由", []InputField{
		NewInputField("标识符", "identifier", "例如: tenant1", true),
		NewInputField("路径模式", "pattern", "例如: /api/{tenant}/users", true),
		NewInputField("标识来源", "identifier_source", "path/json_body/url_param", true),
		NewInputField("目标URL", "target_url", "例如: http://backend:8080", true),
	})
	m.state = StateForm
}

// showEditForm shows the edit form
func (m *RoutesModel) showEditForm() {
	if m.selected == nil {
		return
	}
	m.form = NewForm("编辑路由", []InputField{
		NewInputField("标识符", "identifier", "", true),
		NewInputField("路径模式", "pattern", "", true),
		NewInputField("标识来源", "identifier_source", "path/json_body/url_param", true),
		NewInputField("目标URL", "target_url", "", true),
	})
	m.form.SetValue("identifier", m.selected.Identifier)
	m.form.SetValue("pattern", m.selected.Pattern)
	m.form.SetValue("identifier_source", m.selected.IdentifierSource)
	m.form.SetValue("target_url", m.selected.TargetUrl)
	m.state = StateForm
}

// saveRoute saves the route
func (m *RoutesModel) saveRoute() tea.Cmd {
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
			// Update existing
			params := &db.UpdateRouteParams{
				Pattern:          values["pattern"],
				Identifier:       values["identifier"],
				IdentifierSource: values["identifier_source"],
				TargetUrl:        values["target_url"],
				Enabled:          m.selected.Enabled,
				ID:               m.selected.ID,
			}
			_, err := queries.UpdateRoute(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
			m.selected = nil
		} else {
			// Create new
			params := &db.CreateRouteParams{
				ID:               generateID(),
				Pattern:          values["pattern"],
				Identifier:       values["identifier"],
				IdentifierSource: values["identifier_source"],
				TargetUrl:        values["target_url"],
			}
			_, err := queries.CreateRoute(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
		}

		m.form = nil
		m.state = StateList
		return tea.Batch(m.loadRoutes(), SendSuccess("路由已保存"))()
	}
}

// deleteRoute deletes the selected route
func (m *RoutesModel) deleteRoute() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteRoute(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadRoutes(), SendSuccess("路由已删除"))()
	}
}

// toggleRoute toggles route enabled status
func (m *RoutesModel) toggleRoute(route *db.Route) tea.Cmd {
	return func() tea.Msg {
		newEnabled := int64(1)
		if route.Enabled.Int64 == 1 {
			newEnabled = 0
		}

		params := &db.UpdateRouteParams{
			Pattern:          route.Pattern,
			Identifier:       route.Identifier,
			IdentifierSource: route.IdentifierSource,
			TargetUrl:        route.TargetUrl,
			Enabled:          sql.NullInt64{Int64: newEnabled, Valid: true},
			ID:               route.ID,
		}

		queries := repository.New()
		_, err := queries.UpdateRoute(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadRoutes(), SendSuccess("状态已切换"))()
	}
}