package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	Label string
	Value string
}

type SelectResult struct {
	Value string
	Back  bool // Esc was pressed
}

type selectModel struct {
	title    string
	items    []Item
	filtered []int // indices into items
	cursor   int
	filter   string
	result   SelectResult
	done     bool
	height   int
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func RunSelect(title string, items []Item, height int) (SelectResult, error) {
	if height <= 0 {
		height = 20
	}
	m := selectModel{
		title:  title,
		items:  items,
		height: height,
	}
	m.applyFilter()

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return SelectResult{}, err
	}

	fm := final.(selectModel)
	return fm.result, nil
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.result = SelectResult{Back: true}
			m.done = true
			return m, tea.Quit

		case "esc":
			if m.filter != "" {
				m.filter = ""
				m.cursor = 0
				m.applyFilter()
			} else {
				m.result = SelectResult{Back: true}
				m.done = true
				return m, tea.Quit
			}

		case "enter":
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.result = SelectResult{Value: m.items[idx].Value}
				m.done = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
				m.applyFilter()
			}

		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.cursor = 0
				m.applyFilter()
			}
		}
	}

	return m, nil
}

func (m selectModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n")

	if m.filter != "" {
		b.WriteString(filterStyle.Render("/ "+m.filter))
		b.WriteString("\n")
	}

	// viewport
	start := 0
	visible := m.height
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}

	end := start + visible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		idx := m.filtered[i]
		label := m.items[idx].Label

		if i == m.cursor {
			b.WriteString(cursorStyle.Render(fmt.Sprintf("  > %s", label)))
		} else {
			b.WriteString(dimStyle.Render(fmt.Sprintf("    %s", label)))
		}
		b.WriteString("\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("    No matches"))
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render("  ↑↓ navigate • type to filter • esc back • enter select"))

	return b.String()
}

func (m *selectModel) applyFilter() {
	if m.filter == "" {
		m.filtered = make([]int, len(m.items))
		for i := range m.items {
			m.filtered[i] = i
		}
		return
	}

	q := strings.ToLower(m.filter)
	m.filtered = nil
	for i, item := range m.items {
		if strings.Contains(strings.ToLower(item.Label), q) {
			m.filtered = append(m.filtered, i)
		}
	}
}
