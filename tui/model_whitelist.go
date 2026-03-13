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

// WhitelistModel manages the IP whitelist
type WhitelistModel struct {
	list     list.Model
	entries  []*db.IpWhitelist
	state    AppState
	form     *Form
	selected *db.IpWhitelist
	width    int
	height   int
	keys     KeyMap
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
		entries, err := queries.ListWhitelist(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		m.entries = entries
		return MsgRefresh{}
	}
}

// refreshList refreshes the list
func (m *WhitelistModel) refreshList() {
	items := make([]list.Item, len(m.entries))
	for i, e := range m.entries {
		status := "启用"
		if e.Enabled.Int64 == 0 {
			status = "禁用"
		}
		desc := e.Description.String
		if desc == "" {
			desc = "-"
		}
		items[i] = MenuItem{
			title:       fmt.Sprintf("%s [%s]", e.IpCidr, status),
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
				if len(m.entries) > 0 && m.list.Index() < len(m.entries) {
					return m, m.toggleEntry(m.entries[m.list.Index()])
				}
			}

		case StateForm:
			switch {
			case key.Matches(msg, m.keys.Enter):
				return m, m.saveEntry()
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
	switch m.state {
	case StateList:
		if len(m.entries) == 0 {
			return Header("IP 白名单") + "\n\n" +
				EmptyListMessage("暂无白名单条目，按 'n' 添加") + "\n\n" +
				Help("n 新建", "d 删除", "space 切换状态")
		}
		return m.list.View() + "\n\n" +
			Help("n 新建", "d 删除", "space 切换状态")

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

// toggleEntry toggles entry status
func (m *WhitelistModel) toggleEntry(entry *db.IpWhitelist) tea.Cmd {
	return func() tea.Msg {
		newEnabled := int64(1)
		if entry.Enabled.Int64 == 1 {
			newEnabled = 0
		}

		params := &db.UpdateWhitelistEntryParams{
			IpCidr:      entry.IpCidr,
			Description: entry.Description,
			Enabled:     sql.NullInt64{Int64: newEnabled, Valid: true},
			ID:          entry.ID,
		}

		queries := repository.New()
		_, err := queries.UpdateWhitelistEntry(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadEntries(), SendSuccess("状态已切换"))()
	}
}