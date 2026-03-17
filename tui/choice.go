package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ChoiceField 表示一个互斥选择框组件
type ChoiceField struct {
	Label    string   // 字段标签
	Key      string   // 字段键名
	Options  []string // 可选项值列表（内部值）
	Labels   []string // 可选项显示标签（显示给用户）
	Selected int      // 当前选中项索引
}

// NewChoiceField 创建一个新的选择框
func NewChoiceField(label, key string, options, labels []string) ChoiceField {
	return ChoiceField{
		Label:    label,
		Key:      key,
		Options:  options,
		Labels:   labels,
		Selected: 0,
	}
}

// Update 处理消息并更新选择框状态
func (c ChoiceField) Update(msg tea.Msg) (ChoiceField, tea.Cmd) {
	// 处理空选项列表
	if len(c.Options) == 0 {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := DefaultKeyMap()
		switch {
		case key.Matches(msg, keys.Left):
			// 向左切换选项
			if c.Selected > 0 {
				c.Selected--
			} else {
				// 循环到最后一个选项
				c.Selected = len(c.Options) - 1
			}
		case key.Matches(msg, keys.Right):
			// 向右切换选项
			if c.Selected < len(c.Options)-1 {
				c.Selected++
			} else {
				// 循环到第一个选项
				c.Selected = 0
			}
		}
	}

	return c, nil
}

// View 渲染选择框
func (c ChoiceField) View(focused bool) string {
	var b strings.Builder

	// 渲染选项框内容
	// 自动换行：每行最多显示 4 个选项
	maxOptionsPerLine := 4
	var lines []string
	var currentLine strings.Builder

	for i, label := range c.Labels {
		optionStr := ""
		if i == c.Selected {
			// 选中项使用 ◉ 图标
			optionStr = styleChoiceOptionSelected.Render("◉ " + label)
		} else {
			// 未选中项使用 ○ 图标
			optionStr = styleChoiceOption.Render("○ " + label)
		}

		// 检查当前行是否已有选项
		if currentLine.Len() > 0 {
			// 计算当前行已有多少个选项
			optionsInCurrentLine := (i - 1) % maxOptionsPerLine
			if optionsInCurrentLine == maxOptionsPerLine-1 {
				// 当前行已满，换行
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			} else {
				currentLine.WriteString("    ") // 选项之间的间距
			}
		}
		currentLine.WriteString(optionStr)
	}

	// 添加最后一行
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// 组合所有行
	content := strings.Join(lines, "\n")

	// 选择框样式
	if focused {
		b.WriteString(styleChoiceBoxFocus.Render(content))
	} else {
		b.WriteString(styleChoiceBox.Render(content))
	}

	return b.String()
}

// Value 返回当前选择的内部值
func (c ChoiceField) Value() string {
	if len(c.Options) == 0 || c.Selected < 0 || c.Selected >= len(c.Options) {
		return ""
	}
	return c.Options[c.Selected]
}

// SetValue 根据值设置选中项
func (c *ChoiceField) SetValue(value string) {
	for i, opt := range c.Options {
		if opt == value {
			c.Selected = i
			return
		}
	}
}

// GetLabel 返回字段标签
func (c ChoiceField) GetLabel() string {
	return c.Label
}

// GetKey 返回字段键名
func (c ChoiceField) GetKey() string {
	return c.Key
}

// SetValueByIndex 根据索引设置选中项
func (c *ChoiceField) SetValueByIndex(index int) {
	if index >= 0 && index < len(c.Options) {
		c.Selected = index
	}
}