package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField 定义表单字段接口
type FormField interface {
	// View 渲染字段视图
	View(focused bool) string
	// Update 处理消息更新字段
	Update(msg tea.Msg) (FormField, tea.Cmd)
	// Value 返回字段值
	Value() string
	// SetValue 设置字段值
	SetValue(value string)
	// GetLabel 返回字段标签
	GetLabel() string
	// GetKey 返回字段键名
	GetKey() string
	// IsRequired 返回是否必填
	IsRequired() bool
	// Focus 聚焦字段
	Focus()
	// Blur 取消聚焦
	Blur()
}

// InputField 表示文本输入字段
type InputField struct {
	Label       string
	Key         string
	Input       textinput.Model
	Required    bool
	Placeholder string
}

// NewInputField 创建一个新的输入字段
func NewInputField(label, key, placeholder string, required bool) InputField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Width = 40
	return InputField{
		Label:       label,
		Key:         key,
		Input:       ti,
		Required:    required,
		Placeholder: placeholder,
	}
}

// View 渲染输入字段
func (i InputField) View(focused bool) string {
	var b strings.Builder

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(i.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(i.Label))
	}

	// 必填标记
	if i.Required {
		b.WriteString(styleFieldRequired.Render(" *"))
	}
	b.WriteString("\n")

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染输入框
	if focused {
		b.WriteString(styleInputFocus.Render(i.Input.View()))
	} else {
		b.WriteString(styleInput.Render(i.Input.View()))
	}

	return b.String()
}

// Update 处理消息更新输入字段
func (i *InputField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd
	i.Input, cmd = i.Input.Update(msg)
	return i, cmd
}

// Value 返回输入字段的值
func (i InputField) Value() string {
	return i.Input.Value()
}

// SetValue 设置输入字段的值
func (i *InputField) SetValue(value string) {
	i.Input.SetValue(value)
}

// GetLabel 返回字段标签
func (i InputField) GetLabel() string {
	return i.Label
}

// GetKey 返回字段键名
func (i InputField) GetKey() string {
	return i.Key
}

// IsRequired 返回是否必填
func (i InputField) IsRequired() bool {
	return i.Required
}

// Focus 聚焦输入字段
func (i *InputField) Focus() {
	i.Input.Focus()
}

// Blur 取消聚焦输入字段
func (i *InputField) Blur() {
	i.Input.Blur()
}

// ChoiceFieldWrapper 包装 ChoiceField 以实现 FormField 接口
type ChoiceFieldWrapper struct {
	ChoiceField
}

// NewChoiceFieldWrapper 创建 ChoiceField 包装器（返回指针）
func NewChoiceFieldWrapper(label, key string, options, labels []string) *ChoiceFieldWrapper {
	return &ChoiceFieldWrapper{
		ChoiceField: NewChoiceField(label, key, options, labels),
	}
}

// View 渲染选择字段
func (c ChoiceFieldWrapper) View(focused bool) string {
	var b strings.Builder

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(c.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(c.Label))
	}
	b.WriteString(styleFieldRequired.Render(" *"))
	b.WriteString("\n")

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染选择框
	b.WriteString(c.ChoiceField.View(focused))

	return b.String()
}

// Update 处理消息更新选择字段
func (c *ChoiceFieldWrapper) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd
	c.ChoiceField, cmd = c.ChoiceField.Update(msg)
	return c, cmd
}

// IsRequired 返回是否必填（选择字段总是必填）
func (c ChoiceFieldWrapper) IsRequired() bool {
	return true
}

// Focus 聚焦选择字段（无操作）
func (c *ChoiceFieldWrapper) Focus() {}

// Blur 取消聚焦选择字段（无操作）
func (c *ChoiceFieldWrapper) Blur() {}

// SelectField 表示下拉选择字段
type SelectField struct {
	Label    string   // 字段标签
	Key      string   // 字段键名
	Options  []string // 可选项列表
	Selected int      // 当前选中索引
	Expanded bool     // 是否展开下拉列表
}

// NewSelectField 创建一个新的下拉选择字段
func NewSelectField(label, key string, options []string, defaultIndex int) *SelectField {
	return &SelectField{
		Label:    label,
		Key:      key,
		Options:  options,
		Selected: defaultIndex,
		Expanded: false,
	}
}

// View 渲染下拉选择字段
func (s *SelectField) View(focused bool) string {
	var b strings.Builder

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(s.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(s.Label))
	}
	b.WriteString(styleFieldRequired.Render(" *"))
	b.WriteString("\n")

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 获取当前选中项显示文本
	currentText := ""
	if len(s.Options) > 0 && s.Selected >= 0 && s.Selected < len(s.Options) {
		currentText = s.Options[s.Selected]
	}

	if s.Expanded {
		// 展开状态：显示当前选项 + 下拉列表
		// 当前选中项
		headerText := currentText + "   ▾"
		b.WriteString(styleSelectExpanded.Render(headerText))
		b.WriteString("\n")

		// 渲染选项列表
		for i, opt := range s.Options {
			if focused {
				b.WriteString(styleFocusBar.Render("┃ "))
			} else {
				b.WriteString("  ")
			}
			if i == s.Selected {
				b.WriteString(styleSelectOptionSelected.Render("  ◄ " + opt))
			} else {
				b.WriteString(styleSelectOption.Render("    " + opt))
			}
			b.WriteString("\n")
		}
	} else {
		// 收起状态：显示当前选项按钮
		btnText := currentText + "   ▾"
		if focused {
			b.WriteString(styleSelectBoxFocus.Render(btnText))
		} else {
			b.WriteString(styleSelectBox.Render(btnText))
		}
	}

	return b.String()
}

// Update 处理消息更新下拉选择字段
func (s *SelectField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	// 处理空选项列表
	if len(s.Options) == 0 {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := DefaultKeyMap()
		switch {
		case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Toggle):
			// Enter 或 Space 切换展开/收起状态
			if s.Expanded {
				// 已展开：确认选择并收起
				s.Expanded = false
			} else {
				// 收起：展开下拉列表
				s.Expanded = true
			}
		case key.Matches(msg, keys.Up):
			// 上移选中项（仅在展开时有效）
			if s.Expanded {
				if s.Selected > 0 {
					s.Selected--
				} else {
					// 循环到最后一项
					s.Selected = len(s.Options) - 1
				}
			}
		case key.Matches(msg, keys.Down):
			// 下移选中项（仅在展开时有效）
			if s.Expanded {
				if s.Selected < len(s.Options)-1 {
					s.Selected++
				} else {
					// 循环到第一项
					s.Selected = 0
				}
			}
		case key.Matches(msg, keys.Esc):
			// Esc 收起下拉列表
			s.Expanded = false
		}
	}

	return s, nil
}

// Value 返回当前选中的值
func (s *SelectField) Value() string {
	if len(s.Options) == 0 || s.Selected < 0 || s.Selected >= len(s.Options) {
		return ""
	}
	return s.Options[s.Selected]
}

// SetValue 根据值设置选中项
func (s *SelectField) SetValue(value string) {
	for i, opt := range s.Options {
		if opt == value {
			s.Selected = i
			return
		}
	}
}

// GetLabel 返回字段标签
func (s *SelectField) GetLabel() string {
	return s.Label
}

// GetKey 返回字段键名
func (s *SelectField) GetKey() string {
	return s.Key
}

// IsRequired 返回是否必填（下拉选择总是必填）
func (s *SelectField) IsRequired() bool {
	return true
}

// Focus 聚焦下拉选择字段
func (s *SelectField) Focus() {
	// 无需特殊操作
}

// Blur 取消聚焦下拉选择字段
func (s *SelectField) Blur() {
	// 收起下拉列表
	s.Expanded = false
}

// NumberField 表示数字输入字段
type NumberField struct {
	Label    string
	Key      string
	Input    textinput.Model
	Min, Max int
}

// NewNumberField 创建一个新的数字输入字段
func NewNumberField(label, key string, defaultValue string, min, max int) *NumberField {
	ti := textinput.New()
	ti.SetValue(defaultValue)
	ti.Width = 10
	return &NumberField{
		Label: label,
		Key:   key,
		Input: ti,
		Min:   min,
		Max:   max,
	}
}

// View 渲染数字输入字段
func (n *NumberField) View(focused bool) string {
	var b strings.Builder

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(n.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(n.Label))
	}
	b.WriteString(styleFieldRequired.Render(" *"))
	b.WriteString("\n")

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	if focused {
		b.WriteString(styleInputFocus.Render(n.Input.View()))
	} else {
		b.WriteString(styleInput.Render(n.Input.View()))
	}

	return b.String()
}

// Update 处理消息更新数字输入字段
func (n *NumberField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 只允许数字、退格、删除等控制键
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			if ch >= '0' && ch <= '9' {
				// 允许数字输入
			} else {
				// 忽略其他字符
				return n, nil
			}
		}
	}

	var cmd tea.Cmd
	n.Input, cmd = n.Input.Update(msg)
	return n, cmd
}

// Value 返回数字输入字段的值
func (n *NumberField) Value() string {
	return n.Input.Value()
}

// SetValue 设置数字输入字段的值
func (n *NumberField) SetValue(value string) {
	n.Input.SetValue(value)
}

// GetLabel 返回字段标签
func (n *NumberField) GetLabel() string {
	return n.Label
}

// GetKey 返回字段键名
func (n *NumberField) GetKey() string {
	return n.Key
}

// IsRequired 返回是否必填
func (n *NumberField) IsRequired() bool {
	return true
}

// Focus 聚焦数字输入字段
func (n *NumberField) Focus() {
	n.Input.Focus()
}

// Blur 取消聚焦数字输入字段
func (n *NumberField) Blur() {
	n.Input.Blur()
}

// URLField 表示 URL 复合字段（协议 + 主机 + 端口）
type URLField struct {
	Label      string
	Key        string
	Protocol   *SelectField    // 协议下拉
	Host       textinput.Model // 主机输入
	Port       textinput.Model // 端口输入
	focusIndex int             // 0=协议, 1=主机, 2=端口
}

// NewURLField 创建一个新的 URL 字段
func NewURLField(label, key string) *URLField {
	hostInput := textinput.New()
	hostInput.Placeholder = "主机地址"
	hostInput.Width = 30

	portInput := textinput.New()
	portInput.SetValue("80")
	portInput.Width = 8

	return &URLField{
		Label:    label,
		Key:      key,
		Protocol: NewSelectField("", "", []string{"http", "https"}, 0),
		Host:     hostInput,
		Port:     portInput,
	}
}

// View 渲染 URL 字段（一行：协议下拉 | 主机输入 | 端口输入）
func (u *URLField) View(focused bool) string {
	var b strings.Builder

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(u.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(u.Label))
	}
	b.WriteString(styleFieldRequired.Render(" *"))
	b.WriteString("\n")

	// 聚焦时添加左侧高亮条
	if focused {
		b.WriteString(styleFocusBar.Render("┃ "))
	} else {
		b.WriteString("  ")
	}

	// 渲染一行：协议 | 主机 | 端口
	protocolView := u.Protocol.View(focused && u.focusIndex == 0)
	hostView := u.Host.View()
	portView := u.Port.View()

	// 根据焦点状态渲染不同样式
	if focused {
		if u.focusIndex == 1 {
			hostView = styleInputFocus.Render(hostView)
		} else {
			hostView = styleInput.Render(hostView)
		}
		if u.focusIndex == 2 {
			portView = styleInputFocus.Render(portView)
		} else {
			portView = styleInput.Render(portView)
		}
	} else {
		hostView = styleInput.Render(hostView)
		portView = styleInput.Render(portView)
	}

	// 协议下拉框在展开时需要特殊处理
	if u.Protocol.Expanded {
		// 展开：单独显示协议下拉
		b.WriteString(protocolView)
		b.WriteString("\n")
		// 聚焦条
		if focused {
			b.WriteString(styleFocusBar.Render("┃ "))
		} else {
			b.WriteString("  ")
		}
		b.WriteString(hostView)
		b.WriteString("  ")
		b.WriteString(portView)
	} else {
		// 收起：一行显示
		b.WriteString(protocolView)
		b.WriteString("  ")
		b.WriteString(hostView)
		b.WriteString("  ")
		b.WriteString(portView)
	}

	return b.String()
}

// Update 处理消息更新 URL 字段
func (u *URLField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := DefaultKeyMap()

		// 如果协议下拉展开，优先处理其事件
		if u.Protocol.Expanded {
			updated, c := u.Protocol.Update(msg)
			if sf, ok := updated.(*SelectField); ok {
				u.Protocol = sf
			}
			cmd = c
			// 协议切换时自动调整端口
			if !u.Protocol.Expanded {
				if u.Protocol.Value() == "https" && u.Port.Value() == "80" {
					u.Port.SetValue("443")
				} else if u.Protocol.Value() == "http" && u.Port.Value() == "443" {
					u.Port.SetValue("80")
				}
			}
			return u, cmd
		}

		switch {
		case key.Matches(msg, keys.Left):
			if u.focusIndex > 0 {
				u.focusIndex--
			}
		case key.Matches(msg, keys.Right):
			if u.focusIndex < 2 {
				u.focusIndex++
			}
		case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Toggle):
			if u.focusIndex == 0 {
				u.Protocol.Expanded = true
			}
		default:
			// 根据焦点更新对应的输入框
			switch u.focusIndex {
			case 0:
				updated, c := u.Protocol.Update(msg)
				if sf, ok := updated.(*SelectField); ok {
					u.Protocol = sf
				}
				cmd = c
			case 1:
				u.Host, cmd = u.Host.Update(msg)
			case 2:
				// 只允许数字
				if len(msg.String()) == 1 {
					ch := msg.String()[0]
					if ch < '0' || ch > '9' {
						return u, nil
					}
				}
				u.Port, cmd = u.Port.Update(msg)
			}
		}
	}

	return u, cmd
}

// Value 返回完整的 URL
func (u *URLField) Value() string {
	protocol := u.Protocol.Value()
	host := u.Host.Value()
	port := u.Port.Value()

	if host == "" {
		return ""
	}

	// 默认端口不显示
	if (protocol == "http" && port == "80") || (protocol == "https" && port == "443") {
		return fmt.Sprintf("%s://%s", protocol, host)
	}
	return fmt.Sprintf("%s://%s:%s", protocol, host, port)
}

// SetValue 解析 URL 并设置各子字段
func (u *URLField) SetValue(value string) {
	protocol, host, port := parseTargetURL(value)
	u.Protocol.SetValue(protocol)
	u.Host.SetValue(host)
	u.Port.SetValue(port)
}

// GetLabel 返回字段标签
func (u *URLField) GetLabel() string {
	return u.Label
}

// GetKey 返回字段键名
func (u *URLField) GetKey() string {
	return u.Key
}

// IsRequired 返回是否必填
func (u *URLField) IsRequired() bool {
	return true
}

// Focus 聚焦 URL 字段
func (u *URLField) Focus() {
	u.focusIndex = 0
}

// Blur 取消聚焦 URL 字段
func (u *URLField) Blur() {
	u.Protocol.Expanded = false
}

// Form 表示一个包含多种字段类型的表单
type Form struct {
	title           string
	fields          []FormField
	focusIndex      int
	err             error
	width           int
	height          int
	keys            KeyMap
	visibilityRules map[string]func(*Form) bool // 字段可见性规则
}

// NewForm 创建一个新的表单（向后兼容，接受 InputField 切片）
func NewForm(title string, inputFields []InputField) *Form {
	fields := make([]FormField, len(inputFields))
	for i, f := range inputFields {
		fields[i] = &InputField{
			Label:       f.Label,
			Key:         f.Key,
			Input:       f.Input,
			Required:    f.Required,
			Placeholder: f.Placeholder,
		}
	}
	return NewFormWithFields(title, fields)
}

// NewFormWithFields 创建一个包含多种字段类型的表单
func NewFormWithFields(title string, fields []FormField) *Form {
	f := &Form{
		title:  title,
		fields: fields,
		keys:   DefaultKeyMap(),
	}
	if len(f.fields) > 0 {
		f.fields[0].Focus()
	}
	return f
}

// SetVisibilityRule 设置字段可见性规则
// 当规则函数返回 true 时字段可见，返回 false 时隐藏
func (f *Form) SetVisibilityRule(key string, rule func(*Form) bool) {
	if f.visibilityRules == nil {
		f.visibilityRules = make(map[string]func(*Form) bool)
	}
	f.visibilityRules[key] = rule
}

// GetFieldValue 获取指定字段的值
func (f *Form) GetFieldValue(key string) string {
	for _, field := range f.fields {
		if field.GetKey() == key {
			return field.Value()
		}
	}
	return ""
}

// visibleFields 返回当前可见的字段列表
// 根据 visibilityRules 过滤隐藏字段
func (f *Form) visibleFields() []FormField {
	var visible []FormField
	for _, field := range f.fields {
		key := field.GetKey()
		if rule, exists := f.visibilityRules[key]; exists {
			if !rule(f) {
				continue // 隐藏此字段
			}
		}
		visible = append(visible, field)
	}
	return visible
}

// Init 初始化表单
func (f *Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update 处理表单更新
func (f *Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
	var cmds []tea.Cmd
	visible := f.visibleFields()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.Up):
			if f.focusIndex > 0 {
				visible[f.focusIndex].Blur()
				f.focusIndex--
				visible[f.focusIndex].Focus()
			}
		case key.Matches(msg, f.keys.Down, f.keys.Tab):
			if f.focusIndex < len(visible)-1 {
				visible[f.focusIndex].Blur()
				f.focusIndex++
				visible[f.focusIndex].Focus()
			}
		case key.Matches(msg, f.keys.Enter):
			// 验证并提交
			return f, nil
		}
	}

	// 更新当前聚焦的可见字段
	var cmd tea.Cmd
	visible[f.focusIndex], cmd = visible[f.focusIndex].Update(msg)

	// 同步更新到原始字段列表
	// 找到可见字段在原始字段列表中的索引并更新
	visibleKey := visible[f.focusIndex].GetKey()
	for i, field := range f.fields {
		if field.GetKey() == visibleKey {
			f.fields[i] = visible[f.focusIndex]
			break
		}
	}
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

// View 渲染表单
func (f *Form) View() string {
	var b strings.Builder
	visible := f.visibleFields()

	// 渲染表单标题
	titleLine := styleFormTitle.Render("╭─ " + f.title + " ─╮")
	b.WriteString(titleLine)
	b.WriteString("\n\n")

	// 渲染字段
	for i, field := range visible {
		b.WriteString(field.View(i == f.focusIndex))
		b.WriteString("\n")

		// 添加字段分隔线（最后一个字段前不加）
		if i < len(visible)-1 {
			b.WriteString(styleFormDivider.Render("  ├──────────────────"))
			b.WriteString("\n\n")
		}
	}

	// 渲染底部分隔线
	b.WriteString("\n")
	b.WriteString(styleFormDivider.Render("╰──────────────────────────────╯"))
	b.WriteString("\n\n")

	// 渲染帮助提示
	b.WriteString(styleFormHelp.Render("  ↑↓ 导航   ←→ 切换选项   Enter 确认   Esc 取消"))

	return b.String()
}

// Values 返回表单所有字段的值
func (f *Form) Values() map[string]string {
	values := make(map[string]string)
	for _, field := range f.fields {
		values[field.GetKey()] = field.Value()
	}
	return values
}

// SetValue 设置指定字段的值
func (f *Form) SetValue(key, value string) {
	for _, field := range f.fields {
		if field.GetKey() == key {
			field.SetValue(value)
			return
		}
	}
}

// SetError 设置错误信息
func (f *Form) SetError(err error) {
	f.err = err
}

// Validate 验证必填字段
func (f *Form) Validate() error {
	for _, field := range f.fields {
		if field.IsRequired() && field.Value() == "" {
			return fmt.Errorf("%s 是必填项", field.GetLabel())
		}
	}
	return nil
}