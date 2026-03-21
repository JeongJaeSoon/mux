package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/JeongJaeSoon/mux/internal/config"
	"github.com/JeongJaeSoon/mux/internal/project"
	"github.com/JeongJaeSoon/mux/internal/tmux"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "mux",
	Short: "tmux session manager",
	Long:  "Interactive tmux session manager with TUI",
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	// If there are existing sessions, show picker
	if len(sessions) > 0 {
		options := make([]huh.Option[string], 0, len(sessions)+1)
		for _, s := range sessions {
			options = append(options, huh.NewOption(s, s))
		}
		options = append(options, huh.NewOption("+ New Session", "__new__"))

		var choice string
		err := huh.NewSelect[string]().
			Title("tmux sessions").
			Options(options...).
			Value(&choice).
			Run()
		if err != nil {
			return nil // user cancelled
		}

		if choice != "__new__" {
			return tmux.Attach(choice)
		}
	}

	// Project selection
	return selectAndCreateSession()
}

func selectAndCreateSession() error {
	if _, err := os.Stat(cfg.ReposDir); os.IsNotExist(err) {
		return fmt.Errorf("repos directory not found: %s", cfg.ReposDir)
	}

	projects, err := project.Discover(cfg.ReposDir)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("no projects found in %s", cfg.ReposDir)
	}

	options := make([]huh.Option[int], len(projects))
	for i, p := range projects {
		options[i] = huh.NewOption(p.Name, i)
	}

	var idx int
	err = huh.NewSelect[int]().
		Title("Select project").
		Options(options...).
		Value(&idx).
		Height(20).
		Run()
	if err != nil {
		return nil // user cancelled
	}

	selected := projects[idx]
	branch := project.DetectBranch(selected.Path)
	basename := filepath.Base(selected.Path)
	sessionName := project.SessionName(basename, branch)

	// If session already exists, just attach
	if tmux.SessionExists(sessionName) {
		return tmux.Attach(sessionName)
	}

	if err := tmux.NewSession(sessionName, selected.Path); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return tmux.Attach(sessionName)
}
