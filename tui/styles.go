// Package tui provides the terminal user interface for Prism.
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define colors - 柔和现代配色方案
var (
	// 基础颜色
	colorPrimary = lipgloss.Color("#7D56F4") // 紫色
	colorAccent  = lipgloss.Color("#10B981") // 绿色
	colorError   = lipgloss.Color("#FF6B8A") // 红色
	colorText    = lipgloss.Color("#FAFAFA") // 主文字颜色
	colorMuted   = lipgloss.Color("#6B7280") // 灰色文字
	colorBorder  = lipgloss.Color("#3C3C3C") // 边框颜色

	// 标题样式
	styleTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	// Tab 样式 - 简洁风格
	styleTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	styleTabActive = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorText).
			Background(colorPrimary).
			Bold(true)

	// Tab 底部分隔线样式
	styleTabDivider = lipgloss.NewStyle().
				Foreground(colorPrimary)

	styleHeader = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	// 列表项样式
	styleItem = lipgloss.NewStyle().
			PaddingLeft(2)

	styleItemSelected = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorAccent).
				Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	// 状态样式
	styleError = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	// 输入框样式
	styleInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(50)

	styleInputFocus = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Width(50)

	// 对话框样式
	styleDialog = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	styleDialogError = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorError).
				Padding(1, 2)

	// Choice Box 样式
	styleChoiceBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(55)

	styleChoiceBoxFocus = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Width(55)

	styleChoiceOption = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleChoiceOptionSelected = lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true)

	// Select 下拉列表样式
	styleSelectBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(12)

	styleSelectBoxFocus = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Width(12)

	styleSelectExpanded = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	styleSelectOption = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleSelectOptionSelected = lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true)

	// 列表描述样式
	styleCardDesc = lipgloss.NewStyle().
			Foreground(colorMuted)

	// 状态徽章样式
	styleBadgeEnabled = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))

	styleBadgeDisabled = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B8A"))

	// 空状态样式
	styleEmptyState = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)

	// 表单样式
	styleFormTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	styleFormDivider = lipgloss.NewStyle().
				Foreground(colorBorder)

	// 字段标签样式
	styleFieldLabel = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	styleFieldLabelFocused = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	styleFieldRequired = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	// 聚焦状态的高亮条
	styleFocusBar = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// 帮助提示样式
	styleFormHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1).
			MarginTop(1)
)

// Tab names
const (
	TabRoutes    = "路由"
	TabWhitelist = "白名单"
	TabAPIKeys   = "API Keys"
	TabTLS       = "TLS"
)

// MenuItem represents a menu item
type MenuItem struct {
	title       string
	description string
}

// FilterValue implements list.Item interface
func (m MenuItem) FilterValue() string {
	return m.title
}

// Title implements list.DefaultItem interface
func (m MenuItem) Title() string {
	return m.title
}

// Description implements list.DefaultItem interface
func (m MenuItem) Description() string {
	return m.description
}

// KeyMap defines key bindings
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Esc      key.Binding
	Quit     key.Binding
	Help     key.Binding
	New      key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Toggle   key.Binding
	Save     key.Binding
	Cancel   key.Binding
}

// DefaultKeyMap returns default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "上移"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "下移"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "左移"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "右移"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "确认"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "下一项"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "上一项"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "返回"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "退出"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "帮助"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "新建"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "编辑"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "删除"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "切换"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "保存"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "取消"),
		),
	}
}

// Msg types
type (
	// MsgError is an error message
	MsgError struct{ Err error }
	// MsgSuccess is a success message
	MsgSuccess struct{ Message string }
	// MsgRefresh is a refresh request
	MsgRefresh struct{}
	// MsgTabChange is a tab change request
	MsgTabChange struct{ Tab string }
)

// SendError sends an error message
func SendError(err error) tea.Cmd {
	return func() tea.Msg {
		return MsgError{Err: err}
	}
}

// SendSuccess sends a success message
func SendSuccess(msg string) tea.Cmd {
	return func() tea.Msg {
		return MsgSuccess{Message: msg}
	}
}

// SendRefresh sends a refresh request
func SendRefresh() tea.Msg {
	return MsgRefresh{}
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}