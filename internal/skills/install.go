package skills

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const skillsDir = ".agents/skills"

// SkillsDir returns the absolute path to ~/.agents/skills/.
func SkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, skillsDir), nil
}

// DiscoverSkills finds all directories containing a SKILL.md file
// within the given root directory.
func DiscoverSkills(root string) ([]string, error) {
	var skills []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skills = append(skills, filepath.Dir(path))
		}
		return nil
	})
	return skills, err
}

// Install copies discovered skills from extractedRoot into ~/.agents/skills/
// using the naming convention {org}__{repo}__{skill_dir}.
func Install(prefix, extractedRoot string, only []string) ([]string, error) {
	destBase, err := SkillsDir()
	if err != nil {
		return nil, err
	}
	return InstallTo(destBase, prefix, extractedRoot, only)
}

func InstallTo(destBase, prefix, extractedRoot string, only []string) ([]string, error) {
	skillDirs, err := DiscoverSkills(extractedRoot)
	if err != nil {
		return nil, fmt.Errorf("discovering skills: %w", err)
	}

	if len(skillDirs) == 0 {
		return nil, fmt.Errorf("no skills found (no SKILL.md files)")
	}

	onlySet := make(map[string]bool)
	for _, name := range only {
		onlySet[name] = true
	}

	var installed []string

	for _, skillDir := range skillDirs {
		skillName := filepath.Base(skillDir)

		if len(onlySet) > 0 && !onlySet[skillName] {
			continue
		}

		destName := fmt.Sprintf("%s__%s", prefix, skillName)
		destPath := filepath.Join(destBase, destName)

		if err := os.MkdirAll(destBase, 0755); err != nil {
			return installed, fmt.Errorf("creating skills directory: %w", err)
		}

		// Remove existing skill with the same name
		if err := os.RemoveAll(destPath); err != nil {
			return installed, fmt.Errorf("removing existing skill %s: %w", destName, err)
		}

		if err := copyDir(skillDir, destPath); err != nil {
			return installed, fmt.Errorf("copying skill %s: %w", skillName, err)
		}

		installed = append(installed, destName)
	}

	return installed, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		// Skip hidden directories like .git
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && rel != "." {
			return filepath.SkipDir
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
