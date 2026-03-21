package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
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

	options := make([]huh.Option[string], len(sessions))
	for i, s := range sessions {
		options[i] = huh.NewOption(s, s)
	}

	var selected []string
	err = huh.NewMultiSelect[string]().
		Title("Kill sessions (space to select)").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return nil // user cancelled
	}

	if len(selected) == 0 {
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
