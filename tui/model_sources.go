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

// SourcesModel 管理来源列表
type SourcesModel struct {
	list     list.Model
	sources  []*db.Source
	state    AppState
	form     *Form
	selected *db.Source
	width    int
	height   int
	keys     KeyMap
}

// NewSourcesModel 创建一个新的来源模型
func NewSourcesModel() *SourcesModel {
	m := &SourcesModel{
		state: StateList,
		keys:  DefaultKeyMap(),
	}
	m.list = NewList([]list.Item{}, "来源列表", 80, 20)
	return m
}

// Init 初始化模型
func (m *SourcesModel) Init() tea.Cmd {
	return m.loadSources()
}

// SetSize 设置尺寸
func (m *SourcesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// MsgSourcesLoaded 来源加载完成消息
type MsgSourcesLoaded struct {
	Sources []*db.Source
}

// loadSources 从数据库加载来源
func (m *SourcesModel) loadSources() tea.Cmd {
	return func() tea.Msg {
		queries := repository.New()
		if queries == nil {
			return MsgError{Err: fmt.Errorf("queries not initialized")}
		}
		sources, err := queries.ListSources(context.Background())
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgSourcesLoaded{Sources: sources}
	}
}

// refreshList 刷新列表
func (m *SourcesModel) refreshList() {
	items := make([]list.Item, len(m.sources))
	for i, s := range m.sources {
		desc := s.Description.String
		items[i] = MenuItem{
			title:       s.Name,
			description: desc,
		}
	}
	m.list.SetItems(items)
}

// Update 处理消息
func (m *SourcesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MsgSourcesLoaded:
		m.sources = msg.Sources
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
				if len(m.sources) > 0 && m.list.Index() < len(m.sources) {
					m.selected = m.sources[m.list.Index()]
					m.showEditForm()
				}
				return m, nil
			case key.Matches(msg, m.keys.Delete):
				if len(m.sources) > 0 && m.list.Index() < len(m.sources) {
					m.selected = m.sources[m.list.Index()]
					m.state = StateConfirm
				}
				return m, nil
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
					m.selected = nil
					return m, nil
				}
				// 焦点在确认按钮上才保存
				if m.form != nil && m.form.IsConfirmed() {
					return m, m.saveSource()
				}
				return m, nil
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
				return m, m.deleteSource()
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
func (m *SourcesModel) View() string {
	switch m.state {
	case StateList:
		if len(m.sources) == 0 {
			return Header("来源列表") + "\n\n" +
				EmptyListMessage("暂无来源，按 'n' 创建新来源") + "\n\n" +
				Help("n 新建", "e 编辑", "d 删除")
		}
		items := m.list.Items()
		return RenderSimpleList(items, m.list.Index(), m.height-3) +
			"\n" + Help("n 新建", "e 编辑", "d 删除")

	case StateForm:
		if m.form != nil {
			return m.form.View()
		}

	case StateConfirm:
		if m.selected != nil {
			return Box("确认删除",
				fmt.Sprintf("确定要删除来源 '%s' 吗？\n\n此操作将同时删除关联的项目和路由规则。", m.selected.Name),
				true,
			) + "\n\n" + Help("Enter 确认删除", "Esc 取消")
		}
	}

	return ""
}

// showCreateForm 显示创建表单
func (m *SourcesModel) showCreateForm() {
	m.form = NewFormWithFields("创建来源", []FormField{
		&InputField{
			Label:    "名称",
			Key:      "name",
			Input:    newTextInput("例如: weixin"),
			Required: true,
		},
		&InputField{
			Label:    "描述",
			Key:      "description",
			Input:    newTextInput("例如: 微信回调"),
			Required: false,
		},
	})
	m.state = StateForm
}

// showEditForm 显示编辑表单
func (m *SourcesModel) showEditForm() {
	if m.selected == nil {
		return
	}
	m.form = NewFormWithFields("编辑来源", []FormField{
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
	m.state = StateForm
}

// saveSource 保存来源
func (m *SourcesModel) saveSource() tea.Cmd {
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
			// 更新现有来源
			params := &db.UpdateSourceParams{
				Name:        values["name"],
				Description: sql.NullString{String: values["description"], Valid: values["description"] != ""},
				ID:          m.selected.ID,
			}
			_, err := queries.UpdateSource(context.Background(), params)
			if err != nil {
				return MsgError{Err: err}
			}
			m.selected = nil
		} else {
			// 创建新来源
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

		m.form = nil
		m.state = StateList
		return tea.Batch(m.loadSources(), SendSuccess("来源已保存"))()
	}
}

// deleteSource 删除选中的来源
func (m *SourcesModel) deleteSource() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}

		queries := repository.New()
		err := queries.DeleteSource(context.Background(), m.selected.ID)
		if err != nil {
			return MsgError{Err: err}
		}

		m.selected = nil
		return tea.Batch(m.loadSources(), SendSuccess("来源已删除"))()
	}
}

// GetSelectedSource 获取当前选中的来源
func (m *SourcesModel) GetSelectedSource() *db.Source {
	if len(m.sources) > 0 && m.list.Index() < len(m.sources) {
		return m.sources[m.list.Index()]
	}
	return nil
}

// GetSources 获取所有来源
func (m *SourcesModel) GetSources() []*db.Source {
	return m.sources
}

// GetState 获取当前状态
func (m *SourcesModel) GetState() AppState {
	return m.state
}