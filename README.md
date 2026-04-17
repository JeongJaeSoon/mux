# mux

Interactive tmux session manager with a TUI — pick a project, spin up (or reuse) a tmux session, and optionally create a git worktree on the fly.

## Install

### Homebrew (recommended)

```sh
brew tap JeongJaeSoon/tap
brew install mux
```

Upgrade later with:

```sh
brew update && brew upgrade mux
```

### From source

```sh
go install github.com/JeongJaeSoon/mux@latest
```

## Requirements

- [`tmux`](https://github.com/tmux/tmux) on your `PATH`
- `git` (only if you want the worktree flow)

## Quick start

On first run, `mux` asks for the directory that holds your repositories (defaults to `~/conductor/repos`) and saves the config under `~/.config/mux/`.

```sh
mux              # open the interactive session picker
mux list         # list existing tmux sessions
mux clean        # multi-select sessions to kill
mux version      # print version / commit / build date
```

From the picker you can attach to an existing session or create a new one by selecting a project, naming the session, and (for git repos) optionally creating a new worktree under `.worktrees/<name>`.
