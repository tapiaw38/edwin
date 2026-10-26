package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tapiaw38/edwin/internal/modules"
	cmdrunner "github.com/tapiaw38/edwin/internal/platform/commands"
)

func TestFileManagerCommandArgs(t *testing.T) {
	tests := []struct {
		fileManager string
		path        string
		name        string
		args        []string
	}{
		{fileManager: "", path: "/tmp/folder", name: "xdg-open", args: []string{"/tmp/folder"}},
		{fileManager: "ghostty -e spf", path: "/tmp/folder", name: "ghostty", args: []string{"-e", "spf", "/tmp/folder"}},
		{fileManager: "   ", path: "/tmp/folder", name: "xdg-open", args: []string{"/tmp/folder"}},
	}
	for _, test := range tests {
		name, args := fileManagerCommand(test.fileManager, test.path)
		if name != test.name || !reflect.DeepEqual(args, test.args) {
			t.Errorf("fileManagerCommand(%q, %q) = %q, %v; want %q, %v", test.fileManager, test.path, name, args, test.name, test.args)
		}
	}
}

func TestEveryActionKindHasHandler(t *testing.T) {
	for _, kind := range modules.AllActionKinds() {
		if _, ok := actionHandlers[kind]; !ok {
			t.Errorf("ActionKind %q has no handler in actionHandlers", kind)
		}
	}
}

func TestNoHandlerForUnlistedKind(t *testing.T) {
	listed := make(map[modules.ActionKind]bool)
	for _, kind := range modules.AllActionKinds() {
		listed[kind] = true
	}
	for kind := range actionHandlers {
		if !listed[kind] {
			t.Errorf("handler registered for %q, which AllActionKinds() does not list", kind)
		}
	}
}

func TestZeroActionIsNotExecuted(t *testing.T) {
	if executeActionSpec(modules.ActionSpec{}) {
		t.Error("a zero ActionSpec must not report success")
	}
}

func TestUnknownActionFails(t *testing.T) {
	if executeActionSpec(modules.ActionSpec{Kind: modules.ActionKind("nope")}) {
		t.Error("an unknown ActionKind must not report success")
	}
}

func TestCopyToClipboardFailsWhenNoWriterInstalled(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"edwin-no-such-clipboard-tool"}}

	if copyToClipboard("text") {
		t.Error("copyToClipboard must report failure when no clipboard tool exists")
	}
}

func TestCopyActionDoesNotReportSuccessWhenCopyFails(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"edwin-no-such-clipboard-tool"}}

	action := modules.CopyAction("text").WithStatus("Copied!")
	if executeActionSpec(action) {
		t.Error("a failed copy must not report success, or SuccessStatus overwrites the error")
	}
}

func TestOpenTerminalFailsWhenNoTerminalInstalled(t *testing.T) {
	original := terminals
	t.Cleanup(func() { terminals = original })
	terminals = []string{"edwin-no-such-terminal"}

	if openTerminal("true") {
		t.Error("openTerminal must report failure when no terminal exists")
	}
}

func TestTerminalShellCommandKeepsOutputVisibleUntilEnter(t *testing.T) {
	cmd := cmdrunner.Command("sh", "-c", terminalShellCommand("printf output"))
	cmd.Stdin = strings.NewReader("\n")

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("terminal shell command failed: %v", err)
	}
	if got, want := string(output), "output\nCommand exited with status 0. Press Enter to close."; got != want {
		t.Fatalf("terminal shell output = %q, want %q", got, want)
	}
}

func TestTerminalShellCommandFindsUserLocalExecutables(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	localCommand := filepath.Join(localBin, "edwin-local-command")
	if err := os.WriteFile(localCommand, []byte("#!/bin/sh\nprintf local-command\n"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd := cmdrunner.Command("sh", "-c", terminalShellCommand("edwin-local-command"))
	cmd.Env = []string{"HOME=" + home, "PATH=/usr/bin:/bin"}
	cmd.Stdin = strings.NewReader("\n")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("terminal shell command failed: %v", err)
	}
	if !strings.HasPrefix(string(output), "local-command") {
		t.Fatalf("terminal shell output = %q, want local command output", output)
	}
}

func TestTerminalShellCommandFindsHerdrFromUserLocalBin(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}
	herdr := filepath.Join(localBin, "herdr")
	if err := os.WriteFile(herdr, []byte("#!/bin/sh\nprintf herdr-found\n"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd := cmdrunner.Command("sh", "-c", terminalShellCommand("herdr"))
	cmd.Env = []string{"HOME=" + home, "PATH=/usr/bin:/bin"}
	cmd.Stdin = strings.NewReader("\n")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("terminal shell command failed: %v", err)
	}
	if !strings.HasPrefix(string(output), "herdr-found") {
		t.Fatalf("terminal shell output = %q, want herdr output", output)
	}
}

func TestPasteDoesNotFireWhenCopyFails(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"edwin-no-such-clipboard-tool"}}

	if pasteText("snippet body") {
		t.Error("pasteText must not paste stale clipboard content after a failed copy")
	}
}
