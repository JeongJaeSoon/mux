package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JeongJaeSoon/mux/internal/tmux"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Kill selected tmux sessions",
	RunE:  runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No active sessions")
		return nil
	}

	selected, err := runMultiSelect("Kill sessions (space=select, enter=confirm, esc=cancel)", sessions)
	if err != nil || len(selected) == 0 {
		return nil
	}

	for _, s := range selected {
		if err := tmux.KillSession(s); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "failed to kill %s: %v\n", s, err)
			continue
		}
		fmt.Printf("Killed: %s\n", s)
	}

	return nil
}

// multi-select with bubbletea

type multiSelectModel struct {
	title    string
	items    []string
	selected map[int]bool
	cursor   int
	filter   string
	filtered []int
	done     bool
	aborted  bool
}

var (
	msTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	msCheck  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	msDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	msFilter = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func runMultiSelect(title string, items []string) ([]string, error) {
	m := multiSelectModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
	}
	m.applyFilter()

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := final.(multiSelectModel)
	if fm.aborted {
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
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.selected[idx] = !m.selected[idx]
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

func (m multiSelectModel) View() string {
	if m.done || m.aborted {
		return ""
	}

	var b strings.Builder
	b.WriteString(msTitle.Render(m.title) + "\n")

	if m.filter != "" {
		b.WriteString(msFilter.Render("/ "+m.filter) + "\n")
	}

	for vi, idx := range m.filtered {
		label := m.items[idx]
		check := "[ ]"
		if m.selected[idx] {
			check = "[x]"
		}

		if vi == m.cursor {
			b.WriteString(msCheck.Render(fmt.Sprintf("  > %s %s", check, label)) + "\n")
		} else {
			b.WriteString(msDim.Render(fmt.Sprintf("    %s %s", check, label)) + "\n")
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(msDim.Render("    No matches") + "\n")
	}

	b.WriteString(msDim.Render("  ↑↓ navigate • space select • type to filter • enter confirm • esc cancel"))
	return b.String()
}

func (m *multiSelectModel) applyFilter() {
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
		if strings.Contains(strings.ToLower(item), q) {
			m.filtered = append(m.filtered, i)
		}
	}
}
