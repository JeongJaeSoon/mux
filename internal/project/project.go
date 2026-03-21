package project

import (
	"io/fs"
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

		name := d.Name()
		if skipDirs[name] {
			return fs.SkipDir
		}

		depth := strings.Count(filepath.Clean(path), string(filepath.Separator)) - baseDepth
		if depth > 3 {
			return fs.SkipDir
		}

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
