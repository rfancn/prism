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

// 全局配置键名
const GlobalConfigKeyWhitelistEnabled = "ip_whitelist_enabled"

// WhitelistModel manages the IP whitelist
type WhitelistModel struct {
	list            list.Model
	entries         []*db.IpWhitelist
	state           AppState
	form            *Form
	selected        *db.IpWhitelist
	width           int
	height          int
	keys            KeyMap
	whitelistEnabled bool  // 全局开关状态
}

// NewWhitelistModel creates a new whitelist model
func NewWhitelistModel() *WhitelistModel {
	m := &WhitelistModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "IP 白名单", 80, 20)
	return m
}

// Init initializes the model
func (m *WhitelistModel) Init() tea.Cmd {
	return m.loadEntries()
}

// SetSize sets the size
func (m *WhitelistModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// loadEntries loads whitelist entries
func (m *WhitelistModel) loadEntries() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}

		// 加载白名单条目
		entries, err := queries.ListWhitelist(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		m.entries = entries

		// 加载全局开关状态
		config, err := queries.GetGlobalConfig(context.Background(), GlobalConfigKeyWhitelistEnabled)
		if err == nil {
			m.whitelistEnabled = (config.Value == "true" || config.Value == "1")
		} else {
			// 默认关闭
			m.whitelistEnabled = false
		}

		return MsgRefresh{}
	}
}

// refreshList refreshes the list
func (m *WhitelistModel) refreshList() {
	items := make([]list.Item, len(m.entries))
	for i, e := range m.entries {
		desc := e.Description.String
		items[i] = MenuItem{
			title:       e.IpCidr,
			description: desc,
		}
	}
	m.list.SetItems(items)
}

// Update handles messages
func (m *WhitelistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case key.Matches(msg, m.keys.Delete):
				if len(m.entries) > 0 && m.list.Index() < len(m.entries) {
					m.selected = m.entries[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
			case key.Matches(msg, m.keys.Toggle):
				return m, m.toggleGlobalSwitch()
			}

		case StateForm:
			switch {
			case key.Matches(msg, m.keys.Enter):
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
					return m, nil
				}
				// 焦点在确认按钮上才保存
				if m.form != nil && m.form.IsConfirmed() {
					return m, m.saveEntry()
				}
				return m, nil
			case key.Matches(msg, m.keys.Esc):
				m.state = StateList
				m.form = nil
				return m, nil
			}

		case StateConfirm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				m.state = StateList
				return m, m.deleteEntry()
			case key.Matches(msg, m.keys.Esc):
				m.state = StateList
				return m, nil
			}
		}
	}

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
func (m *WhitelistModel) View() string {
	// 构建状态指示
	statusText := "已关闭"
	if m.whitelistEnabled {
		statusText = "已启用"
	}
	statusLine := fmt.Sprintf("IP白名单功能: %s", statusText)

	switch m.state {
	case StateList:
		var content string
		if len(m.entries) == 0 {
			content = EmptyListMessage("暂无白名单条目，按 'n' 添加")
		} else {
			// 使用自定义渲染，避免 list 组件的 ANSI 转义序列影响 tab bar
			items := m.list.Items()
			content = RenderSimpleList(items, m.list.Index(), m.height-5)
		}
		return statusLine + "\n" + content + "\n" +
			Help("n 新建", "d 删除", "space 切换功能开关")

	case StateForm:
		if m.form != nil {
			return m.form.View()
		}

	case StateConfirm:
		if m.selected != nil {
			return Box("确认删除",
				fmt.Sprintf("确定要删除白名单条目 '%s' 吗？", m.selected.IpCidr),
				true,
			) + "\n\n" + Help("Enter 确认", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm shows the create form
func (m *WhitelistModel) showCreateForm() {
	m.form = NewForm("添加白名单", []InputField{
		NewInputField("IP/CIDR", "ip_cidr", "例如: 192.168.1.0/24", true),
		NewInputField("描述", "description", "可选", false),
	})
	m.state = StateForm
}

// saveEntry saves the entry
func (m *WhitelistModel) saveEntry() tea.Cmd {
	return func() tea.Msg {
		if m.form == nil {
			return nil
		}

		if err := m.form.Validate(); err != nil {
			return MsgError{Err: err}
		}

		values := m.form.Values()

		params := &db.CreateWhitelistEntryParams{
			ID:          generateID(),
			IpCidr:      values["ip_cidr"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
			Enabled:     sql.NullInt64{Int64: 1, Valid: true},
		}

		queries := repository.New()
		_, err := queries.CreateWhitelistEntry(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		m.form = nil
		m.state = StateList
		return tea.Batch(m.loadEntries(), SendSuccess("白名单已添加"))()
	}
}

// deleteEntry deletes the selected entry
func (m *WhitelistModel) deleteEntry() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteWhitelistEntry(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadEntries(), SendSuccess("白名单已删除"))()
	}
}

// toggleGlobalSwitch toggles global whitelist feature status
func (m *WhitelistModel) toggleGlobalSwitch() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}

		// 切换状态
		newValue := "true"
		if m.whitelistEnabled {
			newValue = "false"
		}

		err := queries.SetGlobalConfig(context.Background(), &db.SetGlobalConfigParams{
			Key:   GlobalConfigKeyWhitelistEnabled,
			Value: newValue,
		})
		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadEntries(), SendSuccess("IP白名单功能已"+map[bool]string{true: "启用", false: "关闭"}[newValue == "true"]))()
	}
}

// GetState 获取当前状态
func (m *WhitelistModel) GetState() AppState {
	return m.state
}
