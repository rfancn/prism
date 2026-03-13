package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	// Custom styles
	l.Styles.Title = styleTitle
	l.Styles.PaginationStyle = styleHelp

	return l
}

// EmptyListMessage renders an empty list message
func EmptyListMessage(message string) string {
	return lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		Render("  " + message)
}