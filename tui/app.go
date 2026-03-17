package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab represents a tab in the UI
type Tab int

const (
	TabSourcesIndex Tab = iota
	TabProjectsIndex
	TabRouteRulesIndex
	TabWhitelistIndex
	TabTLSIndex
)

// Tab names
var tabNames = []string{
	"来源",
	"项目",
	"路由规则",
	"白名单",
	"TLS",
}

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
	sourcesModel    *SourcesModel
	projectsModel   *ProjectsModel
	routeRulesModel *RouteRulesModel
	whitelistModel  *WhitelistModel

	// Messages
	err     error
	success string
}

// NewApp creates a new TUI application
func NewApp() *App {
	app := &App{
		tabs:      tabNames,
		activeTab: TabSourcesIndex,
		state:     StateList,
		keys:      DefaultKeyMap(),
	}

	// Initialize sub-models
	app.sourcesModel = NewSourcesModel()
	app.projectsModel = NewProjectsModel()
	app.routeRulesModel = NewRouteRulesModel()
	app.whitelistModel = NewWhitelistModel()

	return app
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.sourcesModel.Init(),
		a.projectsModel.Init(),
		a.routeRulesModel.Init(),
		a.whitelistModel.Init(),
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
		a.sourcesModel.SetSize(a.width, contentHeight)
		a.projectsModel.SetSize(a.width, contentHeight)
		a.routeRulesModel.SetSize(a.width, contentHeight)
		a.whitelistModel.SetSize(a.width, contentHeight)

	case tea.KeyMsg:
		// 检查当前活动子模型是否处于表单模式
		currentState := a.getSubModelState()

		// 如果当前处于表单模式，让子模型优先处理按键
		// 只有在列表模式下才处理全局键（如 Tab 切换）
		if currentState != StateList {
			// 表单模式下只处理全局退出和返回键
			switch {
			case key.Matches(msg, a.keys.Quit):
				return a, tea.Quit
			case key.Matches(msg, a.keys.Esc):
				a.state = StateList
				a.err = nil
				a.success = ""
			}
		} else {
			// 列表模式下处理全局键
			switch {
			case key.Matches(msg, a.keys.Quit):
				return a, tea.Quit
			case key.Matches(msg, a.keys.Left):
				if a.activeTab > 0 {
					a.activeTab--
					// 同步筛选条件
					a.syncFilters()
				}
			case key.Matches(msg, a.keys.Right):
				if int(a.activeTab) < len(a.tabs)-1 {
					a.activeTab++
					// 同步筛选条件
					a.syncFilters()
				}
			}
		}

	case MsgError:
		a.err = msg.Err

	case MsgSuccess:
		a.success = msg.Message
		a.err = nil
		a.state = StateList

	// 处理所有子模型的加载消息，确保每个模型都能收到
	case MsgSourcesLoaded:
		// 更新 sourcesModel
		a.sourcesModel.sources = msg.Sources
		a.sourcesModel.refreshList()
		// 同时更新 projectsModel 的来源列表
		a.projectsModel.sources = msg.Sources
		a.projectsModel.refreshList()

	case MsgSourcesLoadedForProject:
		// projectsModel 专用的来源加载消息
		a.projectsModel.sources = msg.Sources
		a.projectsModel.refreshList()

	case MsgProjectsLoaded:
		a.projectsModel.projects = msg.Projects
		a.projectsModel.refreshList()

	case MsgProjectsLoadedForRule:
		a.routeRulesModel.projects = msg.Projects
		a.routeRulesModel.refreshList()

	case MsgPluginsLoaded:
		a.routeRulesModel.plugins = msg.Plugins

	case MsgRouteRulesLoaded:
		a.routeRulesModel.rules = msg.Rules
		a.routeRulesModel.refreshList()
	}

	// Update active sub-model
	switch a.activeTab {
	case TabSourcesIndex:
		m, cmd := a.sourcesModel.Update(msg)
		a.sourcesModel = m.(*SourcesModel)
		cmds = append(cmds, cmd)
		// 同步状态
		a.state = a.sourcesModel.GetState()
	case TabProjectsIndex:
		m, cmd := a.projectsModel.Update(msg)
		a.projectsModel = m.(*ProjectsModel)
		cmds = append(cmds, cmd)
		// 同步状态
		a.state = a.projectsModel.GetState()
	case TabRouteRulesIndex:
		m, cmd := a.routeRulesModel.Update(msg)
		a.routeRulesModel = m.(*RouteRulesModel)
		cmds = append(cmds, cmd)
		// 同步状态
		a.state = a.routeRulesModel.GetState()
	case TabWhitelistIndex:
		m, cmd := a.whitelistModel.Update(msg)
		a.whitelistModel = m.(*WhitelistModel)
		cmds = append(cmds, cmd)
		// 同步状态
		a.state = a.whitelistModel.GetState()
	}

	return a, tea.Batch(cmds...)
}

// getSubModelState 获取当前活动子模型的状态
func (a *App) getSubModelState() AppState {
	switch a.activeTab {
	case TabSourcesIndex:
		return a.sourcesModel.GetState()
	case TabProjectsIndex:
		return a.projectsModel.GetState()
	case TabRouteRulesIndex:
		return a.routeRulesModel.GetState()
	case TabWhitelistIndex:
		return a.whitelistModel.GetState()
	default:
		return StateList
	}
}

// syncFilters 同步筛选条件
// 当从项目Tab切换到路由规则Tab时，自动筛选当前选中的项目的规则
func (a *App) syncFilters() {
	switch a.activeTab {
	case TabRouteRulesIndex:
		// 从项目Tab切换到路由规则Tab时，筛选当前选中的项目
		if proj := a.projectsModel.GetSelectedProject(); proj != nil {
			a.routeRulesModel.SetFilterProject(proj)
		} else {
			a.routeRulesModel.SetFilterProject(nil)
		}
	}
}

// View renders the application
func (a *App) View() string {
	var b strings.Builder

	// Tab bar - 始终在顶部渲染
	tabBar := TabBar(a.tabs, int(a.activeTab))
	b.WriteString(tabBar)
	b.WriteString("\n\n")

	// Content area based on active tab
	content := ""
	switch a.activeTab {
	case TabSourcesIndex:
		content = a.sourcesModel.View()
	case TabProjectsIndex:
		content = a.projectsModel.View()
	case TabRouteRulesIndex:
		content = a.routeRulesModel.View()
	case TabWhitelistIndex:
		content = a.whitelistModel.View()
	case TabTLSIndex:
		content = a.tlsView()
	}
	b.WriteString(content)

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