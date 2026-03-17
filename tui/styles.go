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
	colorPrimary = lipgloss.Color("#3B82F6") // 蓝色
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
			Foreground(colorMuted)

	styleItemSelected = lipgloss.NewStyle().
				Foreground(colorText).
				Background(lipgloss.Color("#1E3A5F")).
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

	styleInputFocus = styleInput.Copy().BorderForeground(colorPrimary)

	// 对话框样式
	styleDialog = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	styleDialogError = styleDialog.Copy().BorderForeground(colorError)

	// Choice Box 样式
	styleChoiceBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Width(55)

	styleChoiceBoxFocus = styleChoiceBox.Copy().BorderForeground(colorPrimary)

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
			Width(30)

	styleSelectBoxFocus = styleSelectBox.Copy().BorderForeground(colorPrimary)

	styleSelectExpanded = styleSelectBox.Copy().BorderForeground(colorPrimary)

	styleSelectOption = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleSelectOptionSelected = lipgloss.NewStyle().
					Foreground(colorAccent).
					Bold(true)

	// 列表描述样式
	styleCardDesc = lipgloss.NewStyle().
			Foreground(colorMuted)

	// 空状态样式
	styleEmptyState = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)

	// 表单样式
	styleFormTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

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

	// 帮助提示样式
	styleFormHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1).
			MarginTop(1)
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
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Tab    key.Binding
	Esc    key.Binding
	Quit   key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Toggle key.Binding
}

// DefaultKeyMap returns default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "上移"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "下移"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "左移"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "右移"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "确认"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "下一项"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "返回"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "退出"),
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