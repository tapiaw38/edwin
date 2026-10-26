package history

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRecentPrefersLastUsedThenCount(t *testing.T) {
	entries = map[string]entry{
		"Ghostty": {Count: 66, Used: 100},
		"Claude":  {Count: 17, Used: 200},
		"Cursor":  {Count: 16},
		"Vim":     {Count: 1},
	}
	if got := strings.Join(Recent(), ","); got != "Claude,Ghostty,Cursor,Vim" {
		t.Fatalf("unexpected order: %s", got)
	}
}

func TestRecentCapsAtLimit(t *testing.T) {
	entries = make(map[string]entry)
	for i := 0; i < recentLimit+5; i++ {
		entries["App"+strconv.Itoa(i)] = entry{Count: i}
	}
	if got := len(Recent()); got != recentLimit {
		t.Fatalf("expected %d entries, got %d", recentLimit, got)
	}
}

func TestLoadReadsLegacyCountsFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte(`{"Ghostty":66,"Vim":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	historyPath = path
	entries = make(map[string]entry)
	load()

	if Score("Ghostty") != 66 || Score("Vim") != 1 {
		t.Fatalf("legacy counts not loaded: %v", entries)
	}
	if got := strings.Join(Recent(), ","); got != "Ghostty,Vim" {
		t.Fatalf("unexpected order: %s", got)
	}
}
