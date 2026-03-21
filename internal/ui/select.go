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
	Back  bool
}

// --- styles (shared) ---

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	filterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// --- filtering (shared) ---

func filterIndices(items []string, query string) []int {
	if query == "" {
		indices := make([]int, len(items))
		for i := range items {
			indices[i] = i
		}
		return indices
	}
	q := strings.ToLower(query)
	var indices []int
	for i, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			indices = append(indices, i)
		}
	}
	return indices
}

// ============================================================
// Single Select
// ============================================================

type selectModel struct {
	title    string
	items    []Item
	labels   []string
	filtered []int
	cursor   int
	filter   string
	result   SelectResult
	done     bool
	height   int
}

func RunSelect(title string, items []Item, height int) (SelectResult, error) {
	if height <= 0 {
		height = 20
	}
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label
	}
	m := selectModel{
		title:  title,
		items:  items,
		labels: labels,
		height: height,
	}
	m.filtered = filterIndices(m.labels, "")

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return SelectResult{}, err
	}

	fm, ok := final.(selectModel)
	if !ok {
		return SelectResult{Back: true}, nil
	}
	return fm.result, nil
}

func (m selectModel) Init() tea.Cmd { return nil }

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
				m.filtered = filterIndices(m.labels, "")
			} else {
				m.result = SelectResult{Back: true}
				m.done = true
				return m, tea.Quit
			}
		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
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
				m.filtered = filterIndices(m.labels, m.filter)
			}
		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.cursor = 0
				m.filtered = filterIndices(m.labels, m.filter)
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
	b.WriteString(titleStyle.Render(m.title) + "\n")

	if m.filter != "" {
		b.WriteString(filterStyle.Render("/ "+m.filter) + "\n")
	}

	start := 0
	if m.cursor >= m.height {
		start = m.cursor - m.height + 1
	}
	end := start + m.height
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		idx := m.filtered[i]
		label := m.items[idx].Label
		if i == m.cursor {
			b.WriteString(cursorStyle.Render(fmt.Sprintf("  > %s", label)) + "\n")
		} else {
			b.WriteString(dimStyle.Render(fmt.Sprintf("    %s", label)) + "\n")
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("    No matches") + "\n")
	}

	b.WriteString(dimStyle.Render("  ↑↓ navigate • type to filter • esc back • enter select"))
	return b.String()
}

// ============================================================
// Multi Select
// ============================================================

type multiSelectModel struct {
	title    string
	items    []string
	selected map[int]bool
	filtered []int
	cursor   int
	filter   string
	done     bool
	aborted  bool
}

func RunMultiSelect(title string, items []string) ([]string, error) {
	m := multiSelectModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
	}
	m.filtered = filterIndices(m.items, "")

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm, ok := final.(multiSelectModel)
	if !ok || fm.aborted {
		return nil, nil
	}

	var result []string
	for i, item := range fm.items {
		if fm.selected[i] {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case " ":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				idx := m.filtered[m.cursor]
				m.selected[idx] = !m.selected[idx]
			}
		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
				m.filtered = filterIndices(m.items, m.filter)
			}
		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.cursor = 0
				m.filtered = filterIndices(m.items, m.filter)
			}
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	if m.done || m.aborted {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(m.title) + "\n")

	if m.filter != "" {
		b.WriteString(filterStyle.Render("/ "+m.filter) + "\n")
	}

	for vi, idx := range m.filtered {
		label := m.items[idx]
		check := "[ ]"
		if m.selected[idx] {
			check = "[x]"
		}
		if vi == m.cursor {
			b.WriteString(cursorStyle.Render(fmt.Sprintf("  > %s %s", check, label)) + "\n")
		} else {
			b.WriteString(dimStyle.Render(fmt.Sprintf("    %s %s", check, label)) + "\n")
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("    No matches") + "\n")
	}

	b.WriteString(dimStyle.Render("  ↑↓ navigate • space select • type to filter • enter confirm • esc cancel"))
	return b.String()
}
