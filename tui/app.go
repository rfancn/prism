package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab represents a tab in the UI
type Tab int

const (
	TabRoutesIndex Tab = iota
	TabWhitelistIndex
	TabAPIKeysIndex
	TabTLSIndex
)

// AppState represents the current state of the application
type AppState int

const (
	StateList AppState = iota
	StateForm
	StateConfirm
	StateLoading
)

// App is the main TUI application
type App struct {
	tabs      []string
	activeTab Tab
	state     AppState
	width     int
	height    int
	keys      KeyMap

	// Sub-models
	routesModel    *RoutesModel
	whitelistModel *WhitelistModel
	apikeysModel   *APIKeysModel

	// Messages
	err     error
	success string
}

// NewApp creates a new TUI application
func NewApp() *App {
	app := &App{
		tabs: []string{
			TabRoutes,
			TabWhitelist,
			TabAPIKeys,
			TabTLS,
		},
		activeTab: TabRoutesIndex,
		state:     StateList,
		keys:      DefaultKeyMap(),
	}

	// Initialize sub-models
	app.routesModel = NewRoutesModel()
	app.whitelistModel = NewWhitelistModel()
	app.apikeysModel = NewAPIKeysModel()

	return app
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.routesModel.Init(),
		a.whitelistModel.Init(),
		a.apikeysModel.Init(),
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Update sub-models - 留出 tab bar (2行) + footer (2行) 的空间
		contentHeight := a.height - 4
		if contentHeight < 5 {
			contentHeight = 5
		}
		a.routesModel.SetSize(a.width, contentHeight)
		a.whitelistModel.SetSize(a.width, contentHeight)
		a.apikeysModel.SetSize(a.width, contentHeight)

	case tea.KeyMsg:
		// Global key handling
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Left):
			if a.state == StateList && a.activeTab > 0 {
				a.activeTab--
			}
		case key.Matches(msg, a.keys.Right):
			if a.state == StateList && int(a.activeTab) < len(a.tabs)-1 {
				a.activeTab++
			}
		case key.Matches(msg, a.keys.Esc):
			if a.state != StateList {
				a.state = StateList
				a.err = nil
				a.success = ""
			}
		}

	case MsgError:
		a.err = msg.Err

	case MsgSuccess:
		a.success = msg.Message
		a.err = nil
		a.state = StateList
	}

	// Update active sub-model
	switch a.activeTab {
	case TabRoutesIndex:
		m, cmd := a.routesModel.Update(msg)
		a.routesModel = m.(*RoutesModel)
		cmds = append(cmds, cmd)
	case TabWhitelistIndex:
		m, cmd := a.whitelistModel.Update(msg)
		a.whitelistModel = m.(*WhitelistModel)
		cmds = append(cmds, cmd)
	case TabAPIKeysIndex:
		m, cmd := a.apikeysModel.Update(msg)
		a.apikeysModel = m.(*APIKeysModel)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	var b strings.Builder

	// Tab bar
	b.WriteString(TabBar(a.tabs, int(a.activeTab)))
	b.WriteString("\n\n")

	// Content area based on active tab
	switch a.activeTab {
	case TabRoutesIndex:
		b.WriteString(a.routesModel.View())
	case TabWhitelistIndex:
		b.WriteString(a.whitelistModel.View())
	case TabAPIKeysIndex:
		b.WriteString(a.apikeysModel.View())
	case TabTLSIndex:
		b.WriteString(a.tlsView())
	}

	// Error/Success messages
	if a.err != nil {
		b.WriteString("\n")
		b.WriteString(Error(a.err.Error()))
	}
	if a.success != "" {
		b.WriteString("\n")
		b.WriteString(Success(a.success))
	}

	// Footer help
	b.WriteString("\n")
	b.WriteString(Help("←→ 切换标签", "q 退出"))

	return b.String()
}

func (a *App) tlsView() string {
	return Box("TLS 配置", `
服务端 TLS:
  状态: 未启用

功能将在后续版本实现。
`, false)
}

// Run starts the TUI application
func Run() error {
	app := NewApp()
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}