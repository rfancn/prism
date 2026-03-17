package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ListDelegate is a custom delegate for list items
type ListDelegate struct{}

// Height returns the height of a list item
func (d ListDelegate) Height() int { return 1 }

// Spacing returns the spacing between items
func (d ListDelegate) Spacing() int { return 0 }

// Update handles item updates
func (d ListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a list item
func (d ListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(MenuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.Title())

	if index == m.Index() {
		fmt.Fprint(w, styleItemSelected.Render(str))
	} else {
		fmt.Fprint(w, styleItem.Render(str))
	}
}

// NewList creates a new list with custom styling
func NewList(items []list.Item, title string, width, height int) list.Model {
	l := list.New(items, ListDelegate{}, width, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	l.Styles.PaginationStyle = styleHelp

	return l
}

// EmptyListMessage renders an empty list message
func EmptyListMessage(message string) string {
	return styleEmptyState.Render("  " + message)
}

// RenderSimpleList 渲染简单列表
func RenderSimpleList(items []list.Item, selectedIndex, height int) string {
	var b strings.Builder

	for i, item := range items {
		if i >= height {
			break
		}

		menuItem, ok := item.(MenuItem)
		if !ok {
			continue
		}

		// 构建行内容
		var line strings.Builder

		// 选中项添加竖线指示器
		if i == selectedIndex {
			line.WriteString("⏵ ")
		} else {
			line.WriteString("  ")
		}

		// 渲染标题
		line.WriteString(menuItem.Title())

		// 渲染描述（同一行，小一号字体）
		if menuItem.Description() != "" && menuItem.Description() != "-" {
			line.WriteString("  ")
			line.WriteString(styleCardDesc.Render(menuItem.Description()))
		}

		// 渲染整行
		if i == selectedIndex {
			b.WriteString(styleItemSelected.Render(line.String()))
		} else {
			b.WriteString(styleItem.Render(line.String()))
		}

		b.WriteString("\n")
	}

	return b.String()
}
