package main

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/edwin/internal/modules"
)

type queryContext struct {
	icon    string
	tooltip string
}

func contextForQuery(query string) queryContext {
	lower := strings.ToLower(strings.TrimSpace(query))
	switch {
	case modules.IsNavQuery(query), modules.IsPickQuery(query):
		return queryContext{"folder-open-symbolic", "Browse files"}
	case modules.IsSpotifyQuery(query):
		return queryContext{"audio-x-generic-symbolic", "Spotify"}
	case modules.IsYouTubePlayerQuery(query), modules.IsYouTubeQuery(query):
		return queryContext{"video-x-generic-symbolic", "Video"}
	case modules.IsFileQuery(query):
		return queryContext{"folder-saved-search-symbolic", "Find files"}
	case lower == "m" || strings.HasPrefix(lower, "m ") || lower == "music" || strings.HasPrefix(lower, "music "):
		return queryContext{"folder-music-symbolic", "Music library"}
	case lower == "bm" || strings.HasPrefix(lower, "bm "):
		return queryContext{"user-bookmarks-symbolic", "Bookmarks"}
	case strings.HasPrefix(query, ">"):
		return queryContext{"utilities-terminal-symbolic", "Run command"}
	case lower == "w" || strings.HasPrefix(lower, "w "):
		return queryContext{"window-symbolic", "Windows"}
	case modules.IsWeatherQuery(query):
		return queryContext{"weather-clear-symbolic", "Weather"}
	case lower == "mail" || strings.HasPrefix(lower, "mail ") || lower == "email" || strings.HasPrefix(lower, "email "):
		return queryContext{"mail-unread-symbolic", "Email"}
	case lower == "def" || strings.HasPrefix(lower, "def ") || lower == "define" || strings.HasPrefix(lower, "define "):
		return queryContext{"accessories-dictionary-symbolic", "Dictionary"}
	case strings.HasPrefix(lower, "timer "):
		return queryContext{"alarm-symbolic", "Timer"}
	case strings.HasPrefix(lower, ";") || hasCommand(lower, "snip"):
		return queryContext{"text-x-generic-symbolic", "Snippets"}
	case hasCommand(lower, "rename", "copy", "move", "undo", "buffer", "buf"):
		return queryContext{"folder-symbolic", "File operations"}
	case hasCommand(lower, "recent"):
		return queryContext{"document-open-recent-symbolic", "Recent items"}
	case hasCommand(lower, "large", "type", "lt"):
		return queryContext{"format-text-bold-symbolic", "Large type"}
	case hasCommand(lower, "help", "?"):
		return queryContext{"help-browser-symbolic", "Help"}
	case hasCommand(lower, "stats", "usage"):
		return queryContext{"utilities-system-monitor-symbolic", "Usage statistics"}
	case hasCommand(lower, "sync", "settings sync"):
		return queryContext{"folder-sync-symbolic", "Sync settings"}
	case hasCommand(lower, "clip", "clipboard", "cb", "paste"):
		return queryContext{"edit-paste-symbolic", "Clipboard"}
	case hasCommand(lower, "spell", "spelling"):
		return queryContext{"tools-check-spelling-symbolic", "Spelling"}
	case hasCommand(lower, "emoji"):
		return queryContext{"face-smile-symbolic", "Emoji"}
	case hasCommand(lower, "b64", "b64d", "url", "urld", "hash", "uuid", "epoch"):
		return queryContext{"applications-engineering-symbolic", "Developer tools"}
	case hasCommand(lower, "lock", "sleep", "hibernate", "reboot", "shutdown", "logout", "trash"):
		return queryContext{"system-shutdown-symbolic", "System"}
	case hasCommand(lower, "g", "gh", "wiki", "ddg", "r", "so"):
		return queryContext{"emblem-web-symbolic", "Web search"}
	case lower == "contact" || strings.HasPrefix(lower, "contact ") || lower == "contacts" || strings.HasPrefix(lower, "contacts "):
		return queryContext{"avatar-default-symbolic", "Contacts"}
	case lower == "ssh" || strings.HasPrefix(lower, "ssh "):
		return queryContext{"network-server-symbolic", "SSH"}
	case lower == "kill" || strings.HasPrefix(lower, "kill "):
		return queryContext{"process-stop-symbolic", "Processes"}
	case lower == "ss" || lower == "screenshot":
		return queryContext{"camera-photo-symbolic", "Screenshot"}
	case lower == "pass" || strings.HasPrefix(lower, "pass "):
		return queryContext{"dialog-password-symbolic", "Passwords"}
	case lower == "calc" || strings.HasPrefix(lower, "calc "):
		return queryContext{"accessories-calculator-symbolic", "Calculator"}
	default:
		return queryContext{"system-search-symbolic", "Search"}
	}
}

func hasCommand(query string, names ...string) bool {
	for _, name := range names {
		if query == name || strings.HasPrefix(query, name+" ") {
			return true
		}
	}
	return false
}

func updateQueryContext(query string) {
	if searchEntry == nil {
		return
	}
	context := contextForQuery(query)
	searchEntry.SetIconFromIconName(gtk.EntryIconPrimary, context.icon)
	searchEntry.SetIconTooltipText(gtk.EntryIconPrimary, context.tooltip)
}
