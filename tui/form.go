package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
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

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(i.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(i.Label))
	}

	// 必填标记
	if i.Required {
		b.WriteString(styleFieldRequired.Render("*"))
	}
	b.WriteString("\n")

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

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(c.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(c.Label))
	}
	b.WriteString(styleFieldRequired.Render("*"))
	b.WriteString("\n")

	// 渲染选择框（选项在下方）
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

// IDChoiceField 带 ID 值的横向选择框（显示标签，存储ID）
type IDChoiceField struct {
	ChoiceField
}

// NewIDChoiceField 创建带 ID 值的横向选择框
// options: 实际存储的值（如ID）
// labels: 显示给用户的标签
func NewIDChoiceField(label, key string, options, labels []string, defaultIndex int) *IDChoiceField {
	cf := NewChoiceField(label, key, options, labels)
	cf.Selected = defaultIndex
	return &IDChoiceField{
		ChoiceField: cf,
	}
}

// View 渲染横向选择框
func (c *IDChoiceField) View(focused bool) string {
	var b strings.Builder

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(c.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(c.Label))
	}
	b.WriteString(styleFieldRequired.Render("*"))
	b.WriteString("\n")

	// 渲染选择框（选项在下方）
	b.WriteString(c.ChoiceField.View(focused))

	return b.String()
}

// Update 处理消息更新选择框
func (c *IDChoiceField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd
	c.ChoiceField, cmd = c.ChoiceField.Update(msg)
	return c, cmd
}

// IsRequired 返回是否必填
func (c *IDChoiceField) IsRequired() bool {
	return true
}

// Focus 聚焦选择框（无操作）
func (c *IDChoiceField) Focus() {}

// Blur 取消聚焦选择框（无操作）
func (c *IDChoiceField) Blur() {}

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

	// 获取当前选中项显示文本
	currentText := ""
	if len(s.Options) > 0 && s.Selected >= 0 && s.Selected < len(s.Options) {
		currentText = s.Options[s.Selected]
	}

	if s.Expanded {
		// 展开状态：标签和当前选项在同一行，下拉列表在下方
		if focused {
			b.WriteString(styleFieldLabelFocused.Render(s.Label))
		} else {
			b.WriteString(styleFieldLabel.Render(s.Label))
		}
		b.WriteString(styleFieldRequired.Render("*"))
		b.WriteString(" ")
		headerText := currentText + "▾"
		b.WriteString(styleSelectExpandedHeader.Render(headerText))
		b.WriteString("\n")

		// 渲染选项列表（简洁边框）
		var optionsBuilder strings.Builder
		for i, opt := range s.Options {
			if i == s.Selected {
				optionsBuilder.WriteString(styleSelectOptionSelected.Render("► " + opt))
			} else {
				optionsBuilder.WriteString(styleSelectOption.Render("  " + opt))
			}
			if i < len(s.Options)-1 {
				optionsBuilder.WriteString("\n")
			}
		}
		b.WriteString(styleSelectDropdown.Render(optionsBuilder.String()))
	} else {
		// 收起状态：标签和下拉框在同一行
		if focused {
			b.WriteString(styleFieldLabelFocused.Render(s.Label))
		} else {
			b.WriteString(styleFieldLabel.Render(s.Label))
		}
		b.WriteString(styleFieldRequired.Render("*"))
		b.WriteString(" ")

		btnText := currentText + " ▾"
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

// IDSelectField 带ID值的选择字段（显示标签，存储ID）
type IDSelectField struct {
	SelectField
	options []string // 实际值（ID）
}

// NewIDSelectField 创建带ID值的选择字段
func NewIDSelectField(label, key string, options, labels []string, defaultIndex int) *IDSelectField {
	sf := NewSelectField(label, key, labels, defaultIndex)
	return &IDSelectField{
		SelectField: *sf,
		options:     options,
	}
}

// Update 处理消息更新选择字段
func (s *IDSelectField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	_, cmd := s.SelectField.Update(msg)
	return s, cmd
}

// Value 返回实际选中的选项值（ID）
func (s *IDSelectField) Value() string {
	if len(s.options) == 0 || s.Selected < 0 || s.Selected >= len(s.options) {
		return ""
	}
	return s.options[s.Selected]
}

// SetValue 根据实际值设置选中项
func (s *IDSelectField) SetValue(value string) {
	for i, opt := range s.options {
		if opt == value {
			s.Selected = i
			return
		}
	}
}

// NumberField 表示数字输入字段
type NumberField struct {
	Label string
	Key   string
	Input textinput.Model
}

// View 渲染数字输入字段
func (n *NumberField) View(focused bool) string {
	var b strings.Builder

	// 渲染标签
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(n.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(n.Label))
	}
	b.WriteString(styleFieldRequired.Render("*"))
	b.WriteString("\n")

	// 渲染输入框
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

// TextAreaField 表示多行文本输入字段（用于CEL表达式等）
type TextAreaField struct {
	Label    string
	Key      string
	TextArea textarea.Model
	Required bool
}

// NewTextAreaField 创建一个新的多行文本输入字段
func NewTextAreaField(label, key, placeholder string, required bool) *TextAreaField {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.CharLimit = 2000

	return &TextAreaField{
		Label:    label,
		Key:      key,
		TextArea: ta,
		Required: required,
	}
}

// View 渲染多行文本输入字段
func (t *TextAreaField) View(focused bool) string {
	var b strings.Builder

	// 渲染标签和输入框在同一行开始
	if focused {
		b.WriteString(styleFieldLabelFocused.Render(t.Label))
	} else {
		b.WriteString(styleFieldLabel.Render(t.Label))
	}
	if t.Required {
		b.WriteString(styleFieldRequired.Render("*"))
	}
	b.WriteString("\n")

	// 渲染文本区域
	if focused {
		b.WriteString(styleInputFocus.Render(t.TextArea.View()))
	} else {
		b.WriteString(styleInput.Render(t.TextArea.View()))
	}

	return b.String()
}

// Update 处理消息更新多行文本输入字段
func (t *TextAreaField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd
	t.TextArea, cmd = t.TextArea.Update(msg)

	// 自动调整高度（基于行数）
	lineCount := len(t.TextArea.Value()) + 1
	if t.TextArea.Value() != "" {
		lineCount = strings.Count(t.TextArea.Value(), "\n") + 1
	}
	// 限制最小高度为3，最大高度为10
	newHeight := lineCount + 1 // 额外加一行留白
	if newHeight < 3 {
		newHeight = 3
	}
	if newHeight > 10 {
		newHeight = 10
	}
	t.TextArea.SetHeight(newHeight)

	return t, cmd
}

// Value 返回多行文本输入字段的值
func (t *TextAreaField) Value() string {
	return t.TextArea.Value()
}

// SetValue 设置多行文本输入字段的值
func (t *TextAreaField) SetValue(value string) {
	t.TextArea.SetValue(value)
}

// GetLabel 返回字段标签
func (t *TextAreaField) GetLabel() string {
	return t.Label
}

// GetKey 返回字段键名
func (t *TextAreaField) GetKey() string {
	return t.Key
}

// IsRequired 返回是否必填
func (t *TextAreaField) IsRequired() bool {
	return t.Required
}

// Focus 聚焦多行文本输入字段
func (t *TextAreaField) Focus() {
	t.TextArea.Focus()
}

// Blur 取消聚焦多行文本输入字段
func (t *TextAreaField) Blur() {
	t.TextArea.Blur()
}

// Form 表示一个包含多种字段类型的表单
type Form struct {
	title           string
	fields          []FormField
	focusIndex      int
	focusOnButtons  bool // 焦点是否在按钮上
	selectedButton  int // 0: 确认, 1: 取消
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

	// 确保 focusIndex 在有效范围内
	if f.focusIndex >= len(visible) {
		f.focusIndex = len(visible) - 1
	}
	if f.focusIndex < 0 {
		f.focusIndex = 0
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 检查当前聚焦的字段在原始字段列表中的实际索引
		// 这对于正确判断下拉框状态很重要
		currentVisibleField := visible[f.focusIndex]
		currentFieldKey := currentVisibleField.GetKey()

		// 在原始字段列表中找到该字段，检查其下拉框状态
		isSelectExpanded := false
		for _, field := range f.fields {
			if field.GetKey() == currentFieldKey {
				// 检查 SelectField 及其嵌入类型
				if sf, ok := field.(*SelectField); ok && sf.Expanded {
					isSelectExpanded = true
					break
				}
				// 检查 IDSelectField (嵌入 SelectField)
				if isf, ok := field.(*IDSelectField); ok && isf.Expanded {
					isSelectExpanded = true
					break
				}
			}
		}

		// 如果下拉框展开，让字段自己处理上下键
		if !isSelectExpanded && !f.focusOnButtons {
			// 检查当前聚焦的字段是否需要处理 Enter 键
			// TextAreaField: Enter 用于换行
			// SelectField/IDSelectField: Enter 用于展开/收起下拉框
			_, isTextArea := currentVisibleField.(*TextAreaField)
			_, isSelect := currentVisibleField.(*SelectField)
			_, isIDSelect := currentVisibleField.(*IDSelectField)
			needsEnter := isTextArea || isSelect || isIDSelect

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
				} else {
					// 在最后一个字段，切换到按钮
					visible[f.focusIndex].Blur()
					f.focusOnButtons = true
					f.selectedButton = 0 // 默认选中确认按钮
				}
			case key.Matches(msg, f.keys.Enter):
				// 如果字段需要处理 Enter 键，让事件传递给字段处理
				// 否则移动到下一个字段或按钮区域（不自动提交）
				if !needsEnter {
					// 移动到下一个字段或按钮
					if f.focusIndex < len(visible)-1 {
						visible[f.focusIndex].Blur()
						f.focusIndex++
						visible[f.focusIndex].Focus()
					} else {
						// 在最后一个字段，切换到按钮
						visible[f.focusIndex].Blur()
						f.focusOnButtons = true
						f.selectedButton = 0 // 默认选中确认按钮
					}
				}
			}
		} else if f.focusOnButtons {
			// 焦点在按钮上
			switch {
			case key.Matches(msg, f.keys.Up):
				// 返回到最后一个字段
				f.focusOnButtons = false
				f.focusIndex = len(visible) - 1
				visible[f.focusIndex].Focus()
			case key.Matches(msg, f.keys.Left):
				if f.selectedButton > 0 {
					f.selectedButton--
				}
			case key.Matches(msg, f.keys.Right, f.keys.Tab):
				if f.selectedButton < 1 {
					f.selectedButton++
				}
			// Enter 键不在这里处理，让 handleFormKeys 处理
			}
		}
	}

	// 更新当前聚焦的可见字段（如果焦点不在按钮上）
	if !f.focusOnButtons {
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
	}

	return f, tea.Batch(cmds...)
}

// SetSize 设置表单尺寸
func (f *Form) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// View 渲染表单
func (f *Form) View() string {
	var b strings.Builder
	visible := f.visibleFields()

	// 确保 focusIndex 在有效范围内
	if len(visible) == 0 {
		return ""
	}
	if f.focusIndex >= len(visible) {
		f.focusIndex = len(visible) - 1
	}
	if f.focusIndex < 0 {
		f.focusIndex = 0
	}

	// 计算每个字段的实际高度
	// 用于更精确的滚动计算
	fieldHeights := make([]int, len(visible))
	for i, field := range visible {
		// 估算字段高度
		baseHeight := 2 // 普通字段：1行内容 + 1行间距
		switch field.(type) {
		case *TextAreaField:
			// TextArea: label行 + 多行输入区域
			baseHeight = 1 + 3 // label行 + 默认3行TextArea
		case *SelectField:
			// SelectField 展开时需要更多行
			if sf, ok := field.(*SelectField); ok && sf.Expanded {
				baseHeight = 1 + len(sf.Options) + 1 // label行 + 选项列表 + 边距
			}
		case *IDSelectField:
			// IDSelectField 展开时需要更多行
			if isf, ok := field.(*IDSelectField); ok && isf.Expanded {
				baseHeight = 1 + len(isf.Options) + 1
			}
		case *ChoiceFieldWrapper, *IDChoiceField:
			// ChoiceField: label行 + 选项行(可能有换行)
			baseHeight = 3 // label行 + 选项行 + 间距
		}
		fieldHeights[i] = baseHeight
	}

	// 标题和帮助行的固定高度
	titleLines := 2           // 标题 + 空行
	helpLines := 1            // 底部帮助
	buttonLines := 3          // 按钮（1行） + 空行（2行）
	scrollIndicatorLines := 2 // 滚动指示器（如果有）

	// 计算可用高度
	availableHeight := f.height - titleLines - helpLines - buttonLines
	if availableHeight < 10 {
		availableHeight = 10
	}

	// 计算显示窗口的起始和结束索引
	startIdx := 0
	endIdx := len(visible)

	// 如果总高度超过可用高度，需要滚动
	// 找到焦点字段，计算窗口位置
	// 确保焦点字段在窗口中间位置
	targetStart := f.focusIndex - 1
	if targetStart < 0 {
		targetStart = 0
	}

	// 计算从 targetStart 开始的高度
	windowHeight := 0
	windowEnd := targetStart
	for j := targetStart; j < len(fieldHeights); j++ {
		if windowHeight+fieldHeights[j]+scrollIndicatorLines > availableHeight {
			break
		}
		windowHeight += fieldHeights[j]
		windowEnd = j + 1
	}

	// 如果窗口末尾没有包含焦点字段，调整窗口
	if windowEnd <= f.focusIndex {
		// 从焦点字段开始计算
		startIdx = f.focusIndex
		windowHeight = 0
		windowEnd = f.focusIndex
		for j := f.focusIndex; j < len(fieldHeights); j++ {
			if windowHeight+fieldHeights[j] > availableHeight {
				break
			}
			windowHeight += fieldHeights[j]
			windowEnd = j + 1
		}
		endIdx = windowEnd
	} else {
		startIdx = targetStart
		endIdx = windowEnd
	}

	// 边界调整
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(visible) {
		endIdx = len(visible)
	}

	// 渲染表单标题
	b.WriteString(styleFormTitle.Render(f.title))
	b.WriteString("\n\n")

	// 渲染滚动指示器（如果有隐藏字段在上方）
	if startIdx > 0 {
		b.WriteString(styleHelp.Render("  ↑ 更多字段..."))
		b.WriteString("\n\n")
	}

	// 渲染可见字段
	for i := startIdx; i < endIdx; i++ {
		field := visible[i]
		b.WriteString(field.View(i == f.focusIndex))
		b.WriteString("\n")
	}

	// 渲染滚动指示器（如果有隐藏字段在下方）
	if endIdx < len(visible) {
		b.WriteString(styleHelp.Render("  ↓ 更多字段..."))
		b.WriteString("\n")
	}

	// 渲染按钮
	b.WriteString("\n")
	confirmBtn := " [ 确认 ] "
	cancelBtn := " [ 取消 ] "

	if f.focusOnButtons {
		if f.selectedButton == 0 {
			confirmBtn = styleButtonSelected.Render(" [ 确认 ] ")
			cancelBtn = styleButton.Render(" [ 取消 ] ")
		} else {
			confirmBtn = styleButton.Render(" [ 确认 ] ")
			cancelBtn = styleButtonSelected.Render(" [ 取消 ] ")
		}
	} else {
		confirmBtn = styleButton.Render(" [ 确认 ] ")
		cancelBtn = styleButton.Render(" [ 取消 ] ")
	}

	b.WriteString("  ")
	b.WriteString(confirmBtn)
	b.WriteString("  ")
	b.WriteString(cancelBtn)
	b.WriteString("\n\n")

	// 渲染帮助提示
	// 根据当前聚焦的字段类型显示不同的提示
	if f.focusOnButtons {
		b.WriteString(styleFormHelp.Render("←→ 切换按钮   Enter 确认   Esc 取消"))
	} else {
		currentField := visible[f.focusIndex]

		if _, isTextArea := currentField.(*TextAreaField); isTextArea {
			b.WriteString(styleFormHelp.Render("↑↓ 导航   Enter 换行   Tab 下一字段   Esc 取消"))
		} else if f.HasExpandedSelect() {
			// 下拉列表展开时
			b.WriteString(styleFormHelp.Render("↑↓ 选择   Enter 确认   Esc 收起"))
		} else if _, isSelect := currentField.(*SelectField); isSelect {
			b.WriteString(styleFormHelp.Render("↑↓ 导航   Space 展开列表   Tab 下一字段   Esc 取消"))
		} else if _, isIDSelect := currentField.(*IDSelectField); isIDSelect {
			b.WriteString(styleFormHelp.Render("↑↓ 导航   Space 展开列表   Tab 下一字段   Esc 取消"))
		} else {
			b.WriteString(styleFormHelp.Render("↑↓ 导航   ←→ 切换选项   Tab 下一字段   Esc 取消"))
		}
	}

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

// IsCancelled 检查用户是否选择了取消按钮
func (f *Form) IsCancelled() bool {
	return f.focusOnButtons && f.selectedButton == 1
}

// IsConfirmed 检查用户是否选择了确认按钮
func (f *Form) IsConfirmed() bool {
	return f.focusOnButtons && f.selectedButton == 0
}

// HasExpandedSelect 检查当前聚焦的字段是否是展开的下拉选择框
func (f *Form) HasExpandedSelect() bool {
	visible := f.visibleFields()
	if f.focusIndex < 0 || f.focusIndex >= len(visible) {
		return false
	}

	// 获取当前聚焦字段的 key
	currentFieldKey := visible[f.focusIndex].GetKey()

	// 在原始字段列表中查找该字段，检查其展开状态
	for _, field := range f.fields {
		if field.GetKey() == currentFieldKey {
			// 检查 SelectField
			if sf, ok := field.(*SelectField); ok && sf.Expanded {
				return true
			}
			// 检查 IDSelectField
			if isf, ok := field.(*IDSelectField); ok && isf.Expanded {
				return true
			}
		}
	}

	return false
}

// IsTextAreaFocused 检查当前聚焦的字段是否是TextArea
func (f *Form) IsTextAreaFocused() bool {
	visible := f.visibleFields()
	if f.focusIndex < 0 || f.focusIndex >= len(visible) {
		return false
	}

	_, ok := visible[f.focusIndex].(*TextAreaField)
	return ok
}
