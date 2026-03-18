package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

	if err := os.MkdirAll(destBase, 0755); err != nil {
		return nil, fmt.Errorf("creating skills directory: %w", err)
	}

	skills, err := buildInstallSkills(extractedRoot, skillDirs)
	if err != nil {
		return nil, err
	}

	selected, err := selectSkills(skills, only)
	if err != nil {
		return nil, err
	}

	var installed []string

	for _, skill := range skills {
		if len(selected) > 0 && !selected[skill.rel] {
			continue
		}

		destName := fmt.Sprintf("%s__%s", prefix, skill.installName)
		destPath := filepath.Join(destBase, destName)

		if err := installSkillAtomic(destBase, skill.dir, destPath); err != nil {
			return installed, fmt.Errorf("copying skill %s: %w", skill.rel, err)
		}

		installed = append(installed, destName)
	}

	return installed, nil
}

type installSkill struct {
	dir         string
	rel         string
	selector    string
	base        string
	installName string
}

func buildInstallSkills(root string, skillDirs []string) ([]installSkill, error) {
	skills := make([]installSkill, 0, len(skillDirs))
	installNames := make(map[string]string, len(skillDirs))

	for _, skillDir := range skillDirs {
		rel, err := filepath.Rel(root, skillDir)
		if err != nil {
			return nil, fmt.Errorf("resolving skill path %s: %w", skillDir, err)
		}

		rel = filepath.ToSlash(rel)
		selector, err := canonicalSkillSelector(root, skillDir, rel)
		if err != nil {
			return nil, err
		}

		parts := strings.Split(selector, "/")
		for i, part := range parts {
			parts[i] = url.PathEscape(part)
		}

		installName := strings.Join(parts, "__")
		if previous, exists := installNames[installName]; exists {
			return nil, fmt.Errorf("skill install name collision between %q and %q", previous, rel)
		}
		installNames[installName] = rel

		skills = append(skills, installSkill{
			dir:         skillDir,
			rel:         rel,
			selector:    selector,
			base:        filepath.Base(skillDir),
			installName: installName,
		})
	}

	return skills, nil
}

func selectSkills(skills []installSkill, only []string) (map[string]bool, error) {
	if len(only) == 0 {
		return nil, nil
	}

	selected := make(map[string]bool, len(only))
	for _, rawSelector := range only {
		selector := filepath.ToSlash(rawSelector)
		var matches []installSkill

		for _, skill := range skills {
			switch {
			case strings.Contains(selector, "/") && (skill.selector == selector || skill.rel == selector):
				matches = append(matches, skill)
			case !strings.Contains(selector, "/") && skill.base == selector:
				matches = append(matches, skill)
			}
		}

		if len(matches) > 1 {
			options := make([]string, 0, len(matches))
			for _, match := range matches {
				options = append(options, match.selector)
			}
			return nil, fmt.Errorf("ambiguous --only %q; use one of: %s", rawSelector, strings.Join(options, ", "))
		}
		if len(matches) == 1 {
			selected[matches[0].rel] = true
		}
	}

	return selected, nil
}

func canonicalSkillSelector(root, skillDir, rel string) (string, error) {
	pluginRoot, pluginName, found, err := findPluginContext(root, skillDir)
	if err != nil {
		return "", err
	}
	if !found {
		return trimCollectionRoot(rel), nil
	}

	pluginRelative, err := filepath.Rel(pluginRoot, skillDir)
	if err != nil {
		return "", fmt.Errorf("resolving plugin-relative path for %s: %w", skillDir, err)
	}

	pluginRelative = filepath.ToSlash(pluginRelative)
	pluginRelative = strings.TrimPrefix(pluginRelative, "skills/")
	if pluginRelative == "." || pluginRelative == "" {
		return pluginName, nil
	}

	return pluginName + "/" + pluginRelative, nil
}

func trimCollectionRoot(rel string) string {
	for _, prefix := range []string{"skills/", "skill/"} {
		if strings.HasPrefix(rel, prefix) {
			trimmed := strings.TrimPrefix(rel, prefix)
			if trimmed != "" {
				return trimmed
			}
		}
	}

	return rel
}

func findPluginContext(root, skillDir string) (pluginRoot string, pluginName string, found bool, err error) {
	current := skillDir
	cleanRoot := filepath.Clean(root)

	for {
		manifestPath := filepath.Join(current, ".claude-plugin", "plugin.json")
		if _, statErr := os.Stat(manifestPath); statErr == nil {
			name, readErr := readPluginName(manifestPath)
			if readErr != nil {
				return "", "", false, readErr
			}
			if name == "" {
				name = filepath.Base(current)
			}
			return current, name, true, nil
		} else if !os.IsNotExist(statErr) {
			return "", "", false, fmt.Errorf("checking plugin manifest %s: %w", manifestPath, statErr)
		}

		if filepath.Clean(current) == cleanRoot {
			break
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", "", false, nil
}

func readPluginName(manifestPath string) (string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("reading plugin manifest %s: %w", manifestPath, err)
	}

	var manifest struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("parsing plugin manifest %s: %w", manifestPath, err)
	}

	return manifest.Name, nil
}

func installSkillAtomic(destBase, srcDir, destPath string) error {
	stageDir, err := os.MkdirTemp(destBase, ".agent-buddy-staging-*")
	if err != nil {
		return fmt.Errorf("creating staging directory: %w", err)
	}
	defer os.RemoveAll(stageDir)

	if err := copyDir(srcDir, stageDir); err != nil {
		return err
	}

	backupPath := destPath + ".bak"
	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("removing old backup: %w", err)
	}

	if _, err := os.Stat(destPath); err == nil {
		if err := os.Rename(destPath, backupPath); err != nil {
			return fmt.Errorf("moving existing skill aside: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking existing skill: %w", err)
	}

	if err := os.Rename(stageDir, destPath); err != nil {
		if _, restoreErr := os.Stat(backupPath); restoreErr == nil {
			_ = os.Rename(backupPath, destPath)
		}
		return fmt.Errorf("replacing skill: %w", err)
	}

	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("removing backup: %w", err)
	}

	return nil
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
