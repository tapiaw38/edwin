package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteReplacesFileAndMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := AtomicWrite(path, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("content = %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func TestLoadDoesNotPartiallyApplyInvalidConfig(t *testing.T) {
	original := Current
	defer func() { Current = original }()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, AppName), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ConfigFile("config.yaml"), []byte("width: nope\n"), 0644); err != nil {
		t.Fatal(err)
	}
	Current = defaultConfig
	if err := Load(); err == nil {
		t.Fatal("Load accepted invalid config")
	}
	if Current.Width != defaultConfig.Width {
		t.Fatalf("partial config applied: width=%d", Current.Width)
	}
}
