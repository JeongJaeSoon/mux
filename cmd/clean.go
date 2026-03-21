package cmd

import (
	"fmt"

	"github.com/JeongJaeSoon/mux/internal/tmux"
	"github.com/JeongJaeSoon/mux/internal/ui"
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

	selected, err := ui.RunMultiSelect("Kill sessions (space=select, enter=confirm, esc=cancel)", sessions)
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
