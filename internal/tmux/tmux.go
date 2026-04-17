package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func ListSessions() ([]string, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		// tmux server not running → no sessions
		return nil, nil
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func SessionExists(name string) bool {
	return exec.Command("tmux", "has-session", "-t", "="+name).Run() == nil
}

func NewSession(name, dir string, initialArgs ...string) error {
	args := []string{"new-session", "-d", "-s", name, "-c", dir}
	args = append(args, initialArgs...)
	return exec.Command("tmux", args...).Run()
}

func KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

func InsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// Attach replaces the current process with tmux attach/switch.
func Attach(name string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}

	var args []string
	if InsideTmux() {
		args = []string{"tmux", "switch-client", "-t", name}
	} else {
		args = []string{"tmux", "attach-session", "-t", name}
	}

	return syscall.Exec(tmuxPath, args, os.Environ())
}
