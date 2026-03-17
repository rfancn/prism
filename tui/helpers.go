package tui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

	tabsRow := strings.Join(renderedTabs, "  ")

	// 计算tabs行的实际显示长度（去掉ANSI样式码）
	actualWidth := lipgloss.Width(tabsRow)

	// 底部分隔线，长度与tabs行一致
	divider := styleTabDivider.Render(strings.Repeat("─", actualWidth))

	return tabsRow + "\n" + divider
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

// newTextInput 创建一个预配置的 textinput
func newTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 40
	return ti
}

// parseTargetURL 解析目标 URL，返回协议、主机和端口
// 输入示例: "http://backend:8080" 或 "https://api.example.com"
// 如果 URL 无效或缺少端口，返回默认值
func parseTargetURL(urlStr string) (protocol, host, port string) {
	// 默认值
	protocol = "http"
	host = ""
	port = "80"

	// 处理空字符串
	if urlStr == "" {
		return
	}

	// 解析 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}

	// 提取协议 (scheme)
	if parsedURL.Scheme != "" {
		protocol = parsedURL.Scheme
	}

	// 根据协议设置默认端口
	if protocol == "https" {
		port = "443"
	} else {
		port = "80"
	}

	// 提取主机名 (去掉端口部分)
	hostname := parsedURL.Hostname()
	if hostname != "" {
		host = hostname
	}

	// 提取端口 (如果有)
	if parsedURL.Port() != "" {
		port = parsedURL.Port()
	}

	return
}

// buildTargetURL 构建目标 URL
// 输入: protocol="http", host="backend", port="8080"
// 输出: "http://backend:8080"
func buildTargetURL(protocol, host, port string) string {
	// 处理默认端口，省略端口号
	var portPart string
	switch {
	case protocol == "http" && port == "80":
		portPart = ""
	case protocol == "https" && port == "443":
		portPart = ""
	case port != "":
		portPart = ":" + port
	default:
		portPart = ""
	}

	// 构建完整 URL
	if portPart != "" {
		return fmt.Sprintf("%s://%s%s", protocol, host, portPart)
	}
	return fmt.Sprintf("%s://%s", protocol, host)
}