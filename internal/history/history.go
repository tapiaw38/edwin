package history

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/tapiaw38/edwin/internal/config"
)

const recentLimit = 12

var historyPath = config.DataFile("history.json")

type entry struct {
	Count int   `json:"count"`
	Used  int64 `json:"used"`
}

var entries map[string]entry

func init() {
	entries = make(map[string]entry)
	load()
}

func load() {
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &entries); err == nil {
		return
	}

	legacy := make(map[string]int)
	if json.Unmarshal(data, &legacy) != nil {
		return
	}
	entries = make(map[string]entry, len(legacy))
	for name, count := range legacy {
		entries[name] = entry{Count: count}
	}
}

func save() {
	data, _ := json.Marshal(entries)
	_ = config.AtomicWrite(historyPath, data, 0644)
}

func Record(appName string) {
	e := entries[appName]
	e.Count++
	e.Used = time.Now().Unix()
	entries[appName] = e
	save()
}

func Score(appName string) int {
	return entries[appName].Count
}

func Snapshot() map[string]int {
	out := make(map[string]int, len(entries))
	for name, e := range entries {
		out[name] = e.Count
	}
	return out
}

func Recent() []string {
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		a, b := entries[names[i]], entries[names[j]]
		if a.Used != b.Used {
			return a.Used > b.Used
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return names[i] < names[j]
	})
	if len(names) > recentLimit {
		names = names[:recentLimit]
	}
	return names
}
