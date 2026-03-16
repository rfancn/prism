package tui

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
		// 使用自定义渲染，避免 list 组件的 ANSI 转义序列影响 tab bar
		items := m.list.Items()
		return RenderSimpleList(items, m.list.Index(), m.height-3) +
			"\n" + Help("n 新建", "e 编辑", "d 删除", "space 切换状态", "enter 详情")

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
	m.form = NewFormWithFields("创建路由", []FormField{
		&InputField{
			Label:    "标识符",
			Key:      "identifier",
			Input:    newTextInput("例如: tenant1"),
			Required: true,
		},
		NewChoiceFieldWrapper("标识识别方式", "identifier_source",
			[]string{"path", "url_param"},
			[]string{"Path", "URL Param"}),
		&InputField{
			Label:    "路径模式",
			Key:      "pattern",
			Input:    newTextInput("例如: /api/{tenant}/users"),
			Required: false,
		},
		NewURLField("目标URL", "target_url"),
	})
	// 设置路径模式字段的可见性规则：仅在 Path 模式下显示
	m.form.SetVisibilityRule("pattern", func(f *Form) bool {
		return f.GetFieldValue("identifier_source") == "path"
	})
	m.state = StateForm
}

// showEditForm shows the edit form
func (m *RoutesModel) showEditForm() {
	if m.selected == nil {
		return
	}
	m.form = NewFormWithFields("编辑路由", []FormField{
		&InputField{
			Label:    "标识符",
			Key:      "identifier",
			Input:    newTextInput(""),
			Required: true,
		},
		NewChoiceFieldWrapper("标识识别方式", "identifier_source",
			[]string{"path", "url_param"},
			[]string{"Path", "URL Param"}),
		&InputField{
			Label:    "路径模式",
			Key:      "pattern",
			Input:    newTextInput(""),
			Required: false,
		},
		NewURLField("目标URL", "target_url"),
	})
	// 设置路径模式字段的可见性规则
	m.form.SetVisibilityRule("pattern", func(f *Form) bool {
		return f.GetFieldValue("identifier_source") == "path"
	})
	// 设置表单值
	m.form.SetValue("identifier", m.selected.Identifier)
	m.form.SetValue("pattern", m.selected.Pattern)
	m.form.SetValue("identifier_source", m.selected.IdentifierSource)
	m.form.SetValue("target_url", m.selected.TargetUrl)
	m.state = StateForm
}

// newTextInput 创建一个预配置的 textinput
func newTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 40
	return ti
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

// parseTargetURL 解析目标 URL，返回协议、主机和端口
// 输入示例: "http://backend:8080" 或 "https://api.example.com"
// 如果 URL 无效或缺少端口，返回默认值
func parseTargetURL(urlStr string) (protocol, host, port string) {
	// 默认值
	protocol = "http"
	host = ""
	port = "80"

	// 处理空字符串
	if urlStr == "" {
		return
	}

	// 解析 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}

	// 提取协议 (scheme)
	if parsedURL.Scheme != "" {
		protocol = parsedURL.Scheme
	}

	// 根据协议设置默认端口
	if protocol == "https" {
		port = "443"
	} else {
		port = "80"
	}

	// 提取主机名 (去掉端口部分)
	hostname := parsedURL.Hostname()
	if hostname != "" {
		host = hostname
	}

	// 提取端口 (如果有)
	if parsedURL.Port() != "" {
		port = parsedURL.Port()
	}

	return
}

// buildTargetURL 构建目标 URL
// 输入: protocol="http", host="backend", port="8080"
// 输出: "http://backend:8080"
func buildTargetURL(protocol, host, port string) string {
	// 处理默认端口，省略端口号
	var portPart string
	switch {
	case protocol == "http" && port == "80":
		portPart = ""
	case protocol == "https" && port == "443":
		portPart = ""
	case port != "":
		portPart = ":" + port
	default:
		portPart = ""
	}

	// 构建完整 URL
	if portPart != "" {
		return fmt.Sprintf("%s://%s%s", protocol, host, portPart)
	}
	return fmt.Sprintf("%s://%s", protocol, host)
}