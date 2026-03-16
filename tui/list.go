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
		fmt.Fprint(w, styleItemSelected.Render("▶ "+str))
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

// RenderBadge 渲染状态徽章
func RenderBadge(text string, enabled bool) string {
	if enabled {
		return styleBadgeEnabled.Render("● " + text)
	}
	return styleBadgeDisabled.Render("○ " + text)
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

		// 渲染标题行
		if i == selectedIndex {
			b.WriteString(styleItemSelected.Render("▶ " + menuItem.Title()))
		} else {
			b.WriteString(styleItem.Render("  " + menuItem.Title()))
		}

		// 渲染描述行
		if menuItem.Description() != "" {
			b.WriteString("\n")
			desc := formatDescription(menuItem.Description())
			b.WriteString(styleCardDesc.Render("    " + desc))
		}

		b.WriteString("\n")
	}

	return b.String()
}

// formatDescription 格式化描述信息
func formatDescription(desc string) string {
	if strings.Contains(desc, "[启用]") {
		badge := RenderBadge("启用", true)
		return strings.Replace(desc, "[启用]", badge, 1)
	}
	if strings.Contains(desc, "[禁用]") {
		badge := RenderBadge("禁用", false)
		return strings.Replace(desc, "[禁用]", badge, 1)
	}
	return desc
}