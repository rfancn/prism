package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TabBar renders the tab bar
func TabBar(tabs []string, activeIndex int) string {
	var renderedTabs []string
	for i, tab := range tabs {
		if i == activeIndex {
			renderedTabs = append(renderedTabs, styleTabActive.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, styleTab.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

// Title renders a title
func Title(text string) string {
	return styleTitle.Render(" " + text + " ")
}

// Header renders a header
func Header(text string) string {
	return styleHeader.Render("▸ " + text)
}

// Help renders help text
func Help(keys ...string) string {
	return styleHelp.Render(" " + strings.Join(keys, "  "))
}

// Error renders an error message
func Error(text string) string {
	return styleError.Render("✗ " + text)
}

// Success renders a success message
func Success(text string) string {
	return styleSuccess.Render("✓ " + text)
}

// Box renders a bordered box
func Box(title, content string, isError bool) string {
	var style lipgloss.Style
	if isError {
		style = styleDialogError
	} else {
		style = styleDialog
	}

	var b strings.Builder
	if title != "" {
		b.WriteString(styleHeader.Render(title))
		b.WriteString("\n")
	}
	b.WriteString(content)
	return style.Render(b.String())
}

// StatusBadge renders a status badge
func StatusBadge(enabled bool) string {
	if enabled {
		return styleSuccess.Render("● 启用")
	}
	return styleError.Render("○ 禁用")
}

// Truncate truncates a string to max length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}