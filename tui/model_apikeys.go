package tui

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfancn/prism/autogen/db"
	"github.com/rfancn/prism/repository"
)

// APIKeysModel manages API keys
type APIKeysModel struct {
	list     list.Model
	keys     []*db.ApiKey
	state    AppState
	form     *Form
	selected *db.ApiKey
	width    int
	height   int
	keysMap  KeyMap
}

// NewAPIKeysModel creates a new API keys model
func NewAPIKeysModel() *APIKeysModel {
	m := &APIKeysModel{
		state:   StateList,
		keysMap: DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "API Keys", 80, 20)
	return m
}

// Init initializes the model
func (m *APIKeysModel) Init() tea.Cmd {
	return m.loadKeys()
}

// SetSize sets the size
func (m *APIKeysModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// loadKeys loads API keys
func (m *APIKeysModel) loadKeys() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}
		keys, err := queries.ListAPIKeys(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		m.keys = keys
		return MsgRefresh{}
	}
}

// refreshList refreshes the list
func (m *APIKeysModel) refreshList() {
	items := make([]list.Item, len(m.keys))
	for i, k := range m.keys {
		status := "启用"
		if k.Enabled.Int64 == 0 {
			status = "禁用"
		}
		lastUsed := "从未使用"
		if k.LastUsedAt.Valid {
			lastUsed = k.LastUsedAt.Time.Format("2006-01-02 15:04")
		}
		items[i] = MenuItem{
			title:       fmt.Sprintf("%s [%s]", k.Name, status),
			description: fmt.Sprintf("用户: %s | 最后使用: %s", k.UserID, lastUsed),
		}
	}
	m.list.SetItems(items)
}

// Update handles messages
func (m *APIKeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MsgRefresh:
		m.refreshList()

	case tea.KeyMsg:
		switch m.state {
		case StateList:
			switch {
			case key.Matches(msg, m.keysMap.New):
				m.showCreateForm()
				return m, nil
			case key.Matches(msg, m.keysMap.Delete):
				if len(m.keys) > 0 && m.list.Index() < len(m.keys) {
					m.selected = m.keys[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
			case key.Matches(msg, m.keysMap.Toggle):
				if len(m.keys) > 0 && m.list.Index() < len(m.keys) {
					return m, m.toggleKey(m.keys[m.list.Index()])
				}
			}

		case StateForm:
			switch {
			case key.Matches(msg, m.keysMap.Enter):
				return m, m.saveKey()
			case key.Matches(msg, m.keysMap.Esc):
				m.state = StateList
				m.form = nil
				return m, nil
			}

		case StateConfirm:
			switch {
			case key.Matches(msg, m.keysMap.Enter):
				m.state = StateList
				return m, m.deleteKey()
			case key.Matches(msg, m.keysMap.Esc):
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
func (m *APIKeysModel) View() string {
	switch m.state {
	case StateList:
		if len(m.keys) == 0 {
			return Header("API Keys") + "\n\n" +
				EmptyListMessage("暂无 API Key，按 'n' 生成新 Key") + "\n\n" +
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
				fmt.Sprintf("确定要删除 API Key '%s' 吗？", m.selected.Name),
				true,
			) + "\n\n" + Help("Enter 确认", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm shows the create form
func (m *APIKeysModel) showCreateForm() {
	m.form = NewForm("生成 API Key", []InputField{
		NewInputField("名称", "name", "例如: my-app", true),
		NewInputField("用户ID", "user_id", "用于限流标识", true),
		NewInputField("描述", "description", "可选", false),
	})
	m.state = StateForm
}

// saveKey saves the key
func (m *APIKeysModel) saveKey() tea.Cmd {
	return func() tea.Msg {
		if m.form == nil {
			return nil
		}

		if err := m.form.Validate(); err != nil {
			return MsgError{Err: err}
		}

		values := m.form.Values()

		// Generate random API key
		keyBytes := make([]byte, 16)
		if _, err := rand.Read(keyBytes); err != nil {
			return MsgError{Err: err}
		}
		apiKey := "pr_" + hex.EncodeToString(keyBytes)

		params := &db.CreateAPIKeyParams{
			ID:          generateID(),
			Key:         apiKey,
			Name:        values["name"],
			Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
			UserID:      values["user_id"],
			Enabled:     sql.NullInt64{Int64: 1, Valid: true},
		}

		queries := repository.New()
		_, err := queries.CreateAPIKey(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		m.form = nil
		m.state = StateList
		return tea.Batch(
			m.loadKeys(),
			SendSuccess(fmt.Sprintf("API Key 已生成: %s", apiKey)),
		)()
	}
}

// deleteKey deletes the selected key
func (m *APIKeysModel) deleteKey() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteAPIKey(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadKeys(), SendSuccess("API Key 已删除"))()
	}
}

// toggleKey toggles key status
func (m *APIKeysModel) toggleKey(key *db.ApiKey) tea.Cmd {
	return func() tea.Msg {
		newEnabled := int64(1)
		if key.Enabled.Int64 == 1 {
			newEnabled = 0
		}

		params := &db.UpdateAPIKeyParams{
			Name:        key.Name,
			Description: key.Description,
			UserID:      key.UserID,
			Enabled:     sql.NullInt64{Int64: newEnabled, Valid: true},
			ID:          key.ID,
		}

		queries := repository.New()
		_, err := queries.UpdateAPIKey(context.Background(), params)
		if err != nil {
			return MsgError{Err: err}
		}

		return tea.Batch(m.loadKeys(), SendSuccess("状态已切换"))()
	}
}