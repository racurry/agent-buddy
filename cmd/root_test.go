package cmd

import (
	"slices"
	"testing"
)

func TestRootCommandHasExpectedSubcommands(t *testing.T) {
	var names []string
	for _, command := range rootCmd.Commands() {
		names = append(names, command.Name())
	}

	for _, expected := range []string{"install", "list", "uninstall", "version"} {
		if !slices.Contains(names, expected) {
			t.Fatalf("expected subcommand %q to be registered, got %v", expected, names)
		}
	}
}

func TestInstallCommandRejectsInvalidRepoFormat(t *testing.T) {
	err := installCmd.RunE(installCmd, []string{"not-a-repo"})
	if err == nil {
		t.Fatal("expected invalid repo format error")
	}
	if got := err.Error(); got != `invalid repo format "not-a-repo", expected org/repo` {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestInstallCommandFlags(t *testing.T) {
	installOnly = nil
	installPrefix = ""
	installRef = ""

	if err := installCmd.Flags().Set("only", "pdf,docx"); err != nil {
		t.Fatal(err)
	}
	if err := installCmd.Flags().Set("prefix", "custom"); err != nil {
		t.Fatal(err)
	}
	if err := installCmd.Flags().Set("ref", "v1.2.3"); err != nil {
		t.Fatal(err)
	}

	if len(installOnly) != 2 || installOnly[0] != "pdf" || installOnly[1] != "docx" {
		t.Fatalf("unexpected --only parsing: %v", installOnly)
	}
	if installPrefix != "custom" {
		t.Fatalf("unexpected --prefix value: %q", installPrefix)
	}
	if installRef != "v1.2.3" {
		t.Fatalf("unexpected --ref value: %q", installRef)
	}
}
