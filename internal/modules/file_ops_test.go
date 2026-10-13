package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathDigestChangesWhenFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "copy.txt")
	if err := os.WriteFile(path, []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := pathDigest(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("two"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := pathDigest(path)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("digest did not change")
	}
}

func TestUndoCopyRefusesChangedDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "copy.txt")
	if err := os.WriteFile(path, []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	digest, err := pathDigest(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	undo := &fileUndo{Title: "Copy File", Kind: "delete", Source: path, Digest: digest}
	runUndo(undo)
	if _, err := os.Stat(path); err != nil {
		t.Fatal("changed destination was deleted")
	}
}
