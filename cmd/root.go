package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/JeongJaeSoon/mux/internal/config"
	"github.com/JeongJaeSoon/mux/internal/project"
	"github.com/JeongJaeSoon/mux/internal/tmux"
	"github.com/JeongJaeSoon/mux/internal/ui"
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
	if !config.Exists() {
		if err := runSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
			os.Exit(1)
		}
	}

	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
}

func runSetup() error {
	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, "conductor", "repos")

	var reposDir string
	err := huh.NewInput().
		Title("Setup: repos directory").
		Description("Where are your project repositories?").
		Placeholder(defaultDir).
		Value(&reposDir).
		Run()
	if err != nil {
		return err
	}

	if reposDir == "" {
		reposDir = defaultDir
	}

	if len(reposDir) > 1 && reposDir[:2] == "~/" {
		reposDir = filepath.Join(home, reposDir[2:])
	}

	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", reposDir)
	}

	newCfg := &config.Config{ReposDir: reposDir}
	if err := config.Save(newCfg); err != nil {
		return err
	}

	fmt.Printf("Config saved to %s\n", config.ConfigPath())
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	for {
		action, back := showSessionPicker()
		if back {
			return nil
		}

		if action == "__new__" {
			created, back := newSessionFlow()
			if back {
				continue // back to session picker
			}
			if created {
				return nil // attached via syscall.Exec
			}
			continue
		}

		// attach to existing session
		return tmux.Attach(action)
	}
}

func showSessionPicker() (string, bool) {
	sessions, err := tmux.ListSessions()
	if err != nil || len(sessions) == 0 {
		return "__new__", false
	}

	items := make([]ui.Item, 0, len(sessions)+1)
	for _, s := range sessions {
		items = append(items, ui.Item{Label: s, Value: s})
	}
	items = append(items, ui.Item{Label: "+ New Session", Value: "__new__"})

	result, err := ui.RunSelect("tmux sessions", items, 20)
	if err != nil || result.Back {
		return "", true
	}

	return result.Value, false
}

// newSessionFlow: select project → session name → worktree? → create
func newSessionFlow() (bool, bool) {
	// 1. select project
	projectPath, back := pickProject()
	if back {
		return false, true
	}

	// 2. enter session name
	defaultName := sanitizeSessionName(filepath.Base(projectPath))

	var sessionName string
	err := huh.NewInput().
		Title("Session name").
		Placeholder(defaultName).
		Value(&sessionName).
		Run()
	if err != nil {
		return false, true
	}

	if sessionName == "" {
		sessionName = defaultName
	}
	sessionName = sanitizeSessionName(sessionName)

	// 3. ask about worktree
	workDir := projectPath
	if isGitRepo(projectPath) {
		var createWorktree bool
		err := huh.NewConfirm().
			Title("Create new worktree + branch?").
			Value(&createWorktree).
			Run()
		if err != nil {
			return false, true
		}

		if createWorktree {
			branchName := sessionName
			wtPath := filepath.Join(projectPath, ".worktrees", branchName)
			if err := createGitWorktree(projectPath, wtPath, branchName); err != nil {
				fmt.Fprintf(os.Stderr, "failed to create worktree: %v\n", err)
				return false, false
			}
			fmt.Printf("Worktree created: %s (branch: %s)\n", wtPath, branchName)
			workDir = wtPath
		}
	}

	// 4. create and attach
	if tmux.SessionExists(sessionName) {
		return true, tmux.Attach(sessionName) != nil
	}

	if err := tmux.NewSession(sessionName, workDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create session: %v\n", err)
		return false, false
	}

	return true, tmux.Attach(sessionName) != nil
}

func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func createGitWorktree(repoPath, worktreePath, branch string) error {
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branch, worktreePath)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pickProject() (string, bool) {
	if _, err := os.Stat(cfg.ReposDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "repos directory not found: %s\n", cfg.ReposDir)
		return "", true
	}

	projects, err := project.Discover(cfg.ReposDir)
	if err != nil || len(projects) == 0 {
		fmt.Fprintf(os.Stderr, "no projects found in %s\n", cfg.ReposDir)
		return "", true
	}

	items := make([]ui.Item, len(projects))
	for i, p := range projects {
		items[i] = ui.Item{Label: p.Name, Value: p.Path}
	}

	result, err := ui.RunSelect("Select project (esc=back)", items, 20)
	if err != nil || result.Back {
		return "", true
	}

	return result.Value, false
}

func sanitizeSessionName(name string) string {
	r := strings.NewReplacer(
		".", "-", ":", "-", "!", "-",
		" ", "-", "/", "-",
	)
	return r.Replace(name)
}
