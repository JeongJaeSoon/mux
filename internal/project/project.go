package project

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Project struct {
	Name string // relative path from repos_dir
	Path string // absolute path
}

var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	".cache":       true,
	"__pycache__":  true,
	".venv":        true,
	".terraform":   true,
}

func Discover(reposDir string) ([]Project, error) {
	var projects []Project
	baseDepth := strings.Count(filepath.Clean(reposDir), string(filepath.Separator))

	err := filepath.WalkDir(reposDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() {
			return nil
		}

		// skip hidden/excluded dirs
		name := d.Name()
		if skipDirs[name] {
			return fs.SkipDir
		}

		// maxdepth 3
		depth := strings.Count(filepath.Clean(path), string(filepath.Separator)) - baseDepth
		if depth > 3 {
			return fs.SkipDir
		}

		// skip root itself
		if path == reposDir {
			return nil
		}

		rel, err := filepath.Rel(reposDir, path)
		if err != nil {
			return nil
		}
		projects = append(projects, Project{
			Name: rel,
			Path: path,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

func DetectBranch(projectPath string) string {
	out, err := exec.Command("git", "-C", projectPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "main"
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "main"
	}
	return branch
}

func SessionName(basename, branch string) string {
	name := basename + "/" + branch
	// sanitize characters that are problematic for tmux session names
	replacer := strings.NewReplacer(
		".", "-", ":", "-", "!", "-",
		"$", "-", "`", "-", "\\", "-",
		"\"", "-", "'", "-", ";", "-",
		"|", "-", "&", "-", ">", "-",
		"<", "-", " ", "-",
	)
	return replacer.Replace(name)
}
