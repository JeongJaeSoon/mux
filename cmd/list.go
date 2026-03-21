package cmd

import (
	"fmt"

	"github.com/JeongJaeSoon/mux/internal/tmux"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tmux sessions",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	sessions, err := tmux.ListSessions()
	if err != nil {
		return err
	}

	for _, s := range sessions {
		fmt.Println(s)
	}
	return nil
}
