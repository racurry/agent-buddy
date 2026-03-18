package skills

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// makeSkill creates a fake skill directory with a SKILL.md file.
func makeSkill(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverSkills_FindsSkills(t *testing.T) {
	root := t.TempDir()
	makeSkill(t, root, "pdf", "---\nname: pdf\n---")
	makeSkill(t, root, "docx", "---\nname: docx\n---")

	skills, err := DiscoverSkills(root)
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(skills)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	if filepath.Base(skills[0]) != "docx" {
		t.Errorf("expected docx, got %s", filepath.Base(skills[0]))
	}
	if filepath.Base(skills[1]) != "pdf" {
		t.Errorf("expected pdf, got %s", filepath.Base(skills[1]))
	}
}

func TestDiscoverSkills_IgnoresNonSkillDirs(t *testing.T) {
	root := t.TempDir()
	makeSkill(t, root, "pdf", "---\nname: pdf\n---")
	// Directory without SKILL.md
	os.MkdirAll(filepath.Join(root, "not-a-skill"), 0755)
	os.WriteFile(filepath.Join(root, "not-a-skill", "README.md"), []byte("hi"), 0644)

	skills, err := DiscoverSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
}

func TestDiscoverSkills_Nested(t *testing.T) {
	root := t.TempDir()
	makeSkill(t, filepath.Join(root, "subdir"), "nested-skill", "---\nname: nested-skill\n---")

	skills, err := DiscoverSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if filepath.Base(skills[0]) != "nested-skill" {
		t.Errorf("expected nested-skill, got %s", filepath.Base(skills[0]))
	}
}

func TestDiscoverSkills_Empty(t *testing.T) {
	root := t.TempDir()

	skills, err := DiscoverSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(skills))
	}
}

func TestInstallTo_AllSkills(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---\n# PDF Skill")
	makeSkill(t, src, "docx", "---\nname: docx\n---\n# DOCX Skill")

	installed, err := InstallTo(dest, "org__repo", src, nil)
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(installed)
	if len(installed) != 2 {
		t.Fatalf("expected 2 installed, got %d", len(installed))
	}
	if installed[0] != "org__repo__docx" {
		t.Errorf("expected org__repo__docx, got %s", installed[0])
	}
	if installed[1] != "org__repo__pdf" {
		t.Errorf("expected org__repo__pdf, got %s", installed[1])
	}

	// Verify files exist on disk
	content, err := os.ReadFile(filepath.Join(dest, "org__repo__pdf", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "---\nname: pdf\n---\n# PDF Skill" {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestInstallTo_OnlyFilter(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---")
	makeSkill(t, src, "docx", "---\nname: docx\n---")
	makeSkill(t, src, "xlsx", "---\nname: xlsx\n---")

	installed, err := InstallTo(dest, "org__repo", src, []string{"pdf", "xlsx"})
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(installed)
	if len(installed) != 2 {
		t.Fatalf("expected 2 installed, got %d", len(installed))
	}
	if installed[0] != "org__repo__pdf" {
		t.Errorf("expected org__repo__pdf, got %s", installed[0])
	}
	if installed[1] != "org__repo__xlsx" {
		t.Errorf("expected org__repo__xlsx, got %s", installed[1])
	}

	// docx should not exist
	if _, err := os.Stat(filepath.Join(dest, "org__repo__docx")); !os.IsNotExist(err) {
		t.Error("docx should not have been installed")
	}
}

func TestInstallTo_CustomPrefix(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---")

	installed, err := InstallTo(dest, "custom", src, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(installed) != 1 || installed[0] != "custom__pdf" {
		t.Errorf("expected [custom__pdf], got %v", installed)
	}
}

func TestInstallTo_OverwritesExisting(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---\n# Version 1")

	if _, err := InstallTo(dest, "org__repo", src, nil); err != nil {
		t.Fatal(err)
	}

	// Update the source skill
	os.WriteFile(filepath.Join(src, "pdf", "SKILL.md"), []byte("---\nname: pdf\n---\n# Version 2"), 0644)

	if _, err := InstallTo(dest, "org__repo", src, nil); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "org__repo__pdf", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "---\nname: pdf\n---\n# Version 2" {
		t.Errorf("expected version 2, got: %s", content)
	}
}

func TestInstallTo_NoSkillsError(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	_, err := InstallTo(dest, "org__repo", src, nil)
	if err == nil {
		t.Fatal("expected error for empty repo")
	}
}

func TestInstallTo_CopiesSubdirectories(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---")
	// Add a scripts/ subdirectory
	scriptsDir := filepath.Join(src, "pdf", "scripts")
	os.MkdirAll(scriptsDir, 0755)
	os.WriteFile(filepath.Join(scriptsDir, "extract.py"), []byte("print('hi')"), 0644)

	if _, err := InstallTo(dest, "org__repo", src, nil); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "org__repo__pdf", "scripts", "extract.py"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "print('hi')" {
		t.Errorf("unexpected script content: %s", content)
	}
}

func TestInstallTo_SkipsHiddenDirs(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	makeSkill(t, src, "pdf", "---\nname: pdf\n---")
	// Add a .git directory that should be skipped
	gitDir := filepath.Join(src, "pdf", ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("secret"), 0644)

	if _, err := InstallTo(dest, "org__repo", src, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dest, "org__repo__pdf", ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should have been skipped")
	}
}
