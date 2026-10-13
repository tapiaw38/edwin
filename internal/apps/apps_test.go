package apps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationDirsUsesXDGDataDirs(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("XDG_DATA_DIRS", "/var/lib/flatpak/exports/share:/usr/share")

	dirs := applicationDirs()
	want := "/var/lib/flatpak/exports/share/applications"
	for _, dir := range dirs {
		if dir == want {
			return
		}
	}
	t.Fatalf("missing %s in %v", want, dirs)
}

func TestCleanDesktopExecRemovesDesktopFieldCodes(t *testing.T) {
	execCmd := "/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=com.slack.Slack --file-forwarding com.slack.Slack @@u %U @@"
	got := cleanDesktopExec(execCmd)
	want := "/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=com.slack.Slack --file-forwarding com.slack.Slack"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseDesktopExecPreservesQuotedArguments(t *testing.T) {
	got := parseDesktopExec(`/usr/bin/example --title "My App" --path '/tmp/a b' %U`)
	want := []string{"/usr/bin/example", "--title", "My App", "--path", "/tmp/a b"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestLoadFromDirsDeduplicatesDesktopFileIDsByPriority(t *testing.T) {
	root := t.TempDir()
	userDir := filepath.Join(root, "user", "applications")
	systemDir := filepath.Join(root, "system", "applications")
	writeDesktop := func(dir, relativePath, name string) {
		t.Helper()
		path := filepath.Join(dir, relativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		content := "[Desktop Entry]\nName=" + name + "\nExec=" + strings.ToLower(name) + "\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeDesktop(userDir, "code.desktop", "User VS Code")
	writeDesktop(systemDir, "code.desktop", "System VS Code")
	writeDesktop(userDir, "kde/app.desktop", "Nested User App")
	writeDesktop(systemDir, "kde-app.desktop", "Nested System App")

	got := loadFromDirs([]string{userDir, systemDir})
	if len(got) != 2 {
		t.Fatalf("loaded %d apps, want 2: %#v", len(got), got)
	}
	if got[0].Name != "User VS Code" || got[1].Name != "Nested User App" {
		t.Fatalf("loaded apps %#v, want user entries for both desktop-file-IDs", got)
	}
}

func TestLoadFindsFlatpakSlackWhenInstalled(t *testing.T) {
	if _, err := os.Stat("/var/lib/flatpak/exports/share/applications/com.slack.Slack.desktop"); err != nil {
		t.Skip("Slack flatpak desktop file not installed")
	}

	apps := Load()
	for _, app := range apps {
		if app.Name == "Slack" {
			if !strings.Contains(app.Exec, "com.slack.Slack") {
				t.Fatalf("Slack Exec not cleaned: %q", app.Exec)
			}
			return
		}
	}
	t.Fatalf("Slack not loaded from XDG application dirs")
}
