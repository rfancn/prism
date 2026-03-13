// Package tui provides the terminal user interface for Prism.
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define colors
var (
	// Colors
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#3C3C3C")
	colorAccent    = lipgloss.Color("#04B575")
	colorError     = lipgloss.Color("#FF5F87")
	colorWarning   = lipgloss.Color("#FFC857")
	colorText      = lipgloss.Color("#FAFAFA")
	colorMuted     = lipgloss.Color("#626262")

	// Styles
	styleTitle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	styleTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	styleTabActive = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorText).
			Background(colorPrimary).
			Bold(true)

	styleHeader = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	styleItem = lipgloss.NewStyle().
			PaddingLeft(2)

	styleItemSelected = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorAccent).
				Bold(true)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	styleError = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleInput = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorSecondary).
			Padding(0, 1).
			Width(50)

	styleInputFocus = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Width(50)

	styleButton = lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(colorText).
			Background(colorPrimary)

	styleButtonActive = lipgloss.NewStyle().
				Padding(0, 3).
				Foreground(colorText).
				Background(colorAccent).
				Bold(true)

	styleDialog = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	styleDialogError = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorError).
				Padding(1, 2)
)

// Tab names
const (
	TabRoutes     = "路由"
	TabHeaders    = "Headers"
	TabWhitelist  = "白名单"
	TabAPIKeys    = "API Keys"
	TabRateLimit  = "限流"
	TabTLS        = "TLS"
	TabStats      = "统计"
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