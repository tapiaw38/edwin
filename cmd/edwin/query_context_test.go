package main

import "testing"

func TestContextForQuery(t *testing.T) {
	tests := []struct {
		query string
		icon  string
	}{
		{"", "system-search-symbolic"},
		{"nav ~/Code", "folder-open-symbolic"},
		{"sp radiohead", "audio-x-generic-symbolic"},
		{"m albums", "folder-music-symbolic"},
		{"f report", "folder-saved-search-symbolic"},
		{"> git status", "utilities-terminal-symbolic"},
		{"bm docs", "user-bookmarks-symbolic"},
		{"weather Salta", "weather-clear-symbolic"},
		{"ssh server", "network-server-symbolic"},
		{"clip password", "edit-paste-symbolic"},
		{"emoji fire", "face-smile-symbolic"},
		{"shutdown", "system-shutdown-symbolic"},
	}
	for _, test := range tests {
		if got := contextForQuery(test.query).icon; got != test.icon {
			t.Errorf("contextForQuery(%q) icon = %q, want %q", test.query, got, test.icon)
		}
	}
}
