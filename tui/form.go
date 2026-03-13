package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// InputField represents an input field
type InputField struct {
	Label       string
	Key         string
	Input       textinput.Model
	Required    bool
	Placeholder string
}

// NewInputField creates a new input field
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

// Form represents a form with multiple input fields
type Form struct {
	title      string
	fields     []InputField
	focusIndex int
	err        error
	width      int
	height     int
	keys       KeyMap
}

// NewForm creates a new form
func NewForm(title string, fields []InputField) *Form {
	f := &Form{
		title:  title,
		fields: fields,
		keys:   DefaultKeyMap(),
	}
	if len(f.fields) > 0 {
		f.fields[0].Input.Focus()
	}
	return f
}

// Init initializes the form
func (f *Form) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form updates
func (f *Form) Update(msg tea.Msg) (*Form, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.Up):
			if f.focusIndex > 0 {
				f.fields[f.focusIndex].Input.Blur()
				f.focusIndex--
				f.fields[f.focusIndex].Input.Focus()
			}
		case key.Matches(msg, f.keys.Down, f.keys.Tab):
			if f.focusIndex < len(f.fields)-1 {
				f.fields[f.focusIndex].Input.Blur()
				f.focusIndex++
				f.fields[f.focusIndex].Input.Focus()
			}
		case key.Matches(msg, f.keys.Enter):
			// Validate and submit
			return f, nil
		}
	}

	// Update focused input
	var cmd tea.Cmd
	f.fields[f.focusIndex].Input, cmd = f.fields[f.focusIndex].Input.Update(msg)
	cmds = append(cmds, cmd)

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *Form) View() string {
	var b strings.Builder

	b.WriteString(Header(f.title))
	b.WriteString("\n\n")

	for i, field := range f.fields {
		// Label
		label := field.Label
		if field.Required {
			label += " *"
		}
		if i == f.focusIndex {
			b.WriteString(styleInputFocus.Render(label))
		} else {
			b.WriteString("  " + label)
		}
		b.WriteString("\n")

		// Input
		if i == f.focusIndex {
			b.WriteString(styleInputFocus.Render(field.Input.View()))
		} else {
			b.WriteString(styleInput.Render(field.Input.View()))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(Help("↑↓ 导航", "Enter 确认", "Esc 取消"))

	return b.String()
}

// Values returns the form values as a map
func (f *Form) Values() map[string]string {
	values := make(map[string]string)
	for _, field := range f.fields {
		values[field.Key] = field.Input.Value()
	}
	return values
}

// SetValue sets a field value
func (f *Form) SetValue(key, value string) {
	for i, field := range f.fields {
		if field.Key == key {
			f.fields[i].Input.SetValue(value)
			return
		}
	}
}

// SetError sets an error message
func (f *Form) SetError(err error) {
	f.err = err
}

// Validate validates required fields
func (f *Form) Validate() error {
	for _, field := range f.fields {
		if field.Required && field.Input.Value() == "" {
			return fmt.Errorf("%s 是必填项", field.Label)
		}
	}
	return nil
}