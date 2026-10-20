package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Width            int     `yaml:"width"`
	MaxResults       int     `yaml:"max_results"`
	Theme            string  `yaml:"theme"`
	BackgroundColor  string  `yaml:"background_color"`
	BackgroundAlpha  float64 `yaml:"background_alpha"`
	BorderRadius     int     `yaml:"border_radius"`
	FontSize         int     `yaml:"font_size"`
	TextColor        string  `yaml:"text_color"`
	SelectionColor   string  `yaml:"selection_color"`
	FocusBorderColor string  `yaml:"focus_border_color"`
	EntryBackground  string  `yaml:"entry_background"`
	DescColor        string  `yaml:"desc_color"`

	WebShortcuts map[string]WebShortcut `yaml:"web_shortcuts"`

	ShowIcons     bool   `yaml:"show_icons"`
	IconSize      int    `yaml:"icon_size"`
	MarginTop     int    `yaml:"margin_top"`
	Position      string `yaml:"position"`
	HistoryBoost  int    `yaml:"history_boost"`
	SpellLanguage string `yaml:"spell_language"`

	Hotkey           string `yaml:"hotkey"`
	FileManager      string `yaml:"file_manager"`
	ShortcutModifier string `yaml:"shortcut_modifier"`
}

type WebShortcut struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Icon string `yaml:"icon"`
}

var defaultConfig = Config{
	Width:            600,
	MaxResults:       6,
	Theme:            "edwin",
	BackgroundColor:  "24, 26, 34",
	BackgroundAlpha:  0.97,
	BorderRadius:     12,
	FontSize:         18,
	TextColor:        "#fdf2d6",
	SelectionColor:   "208, 193, 180",
	FocusBorderColor: "245, 215, 140",
	EntryBackground:  "rgba(208, 193, 180, 0.2)",
	DescColor:        "rgba(253, 242, 214, 0.62)",
	ShowIcons:        true,
	IconSize:         24,
	MarginTop:        100,
	Position:         "top",
	HistoryBoost:     3,
	SpellLanguage:    "en",
	Hotkey:           "Alt,space",
	ShortcutModifier: "ctrl+shift",
	WebShortcuts: map[string]WebShortcut{
		"g":    {Name: "Google", URL: "https://www.google.com/search?q=%s", Icon: "web-browser"},
		"yt":   {Name: "YouTube", URL: "https://www.youtube.com/results?search_query=%s", Icon: "youtube"},
		"gh":   {Name: "GitHub", URL: "https://github.com/search?q=%s", Icon: "github"},
		"wiki": {Name: "Wikipedia", URL: "https://en.wikipedia.org/wiki/Special:Search?search=%s", Icon: "wikipedia"},
		"ddg":  {Name: "DuckDuckGo", URL: "https://duckduckgo.com/?q=%s", Icon: "web-browser"},
		"r":    {Name: "Reddit", URL: "https://www.reddit.com/search/?q=%s", Icon: "reddit"},
		"so":   {Name: "Stack Overflow", URL: "https://stackoverflow.com/search?q=%s", Icon: "web-browser"},
	},
}

var Current = defaultConfig

func Load() error {
	configPath := ConfigFile("config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Save()
	}

	var probe struct {
		Theme string `yaml:"theme"`
	}
	yaml.Unmarshal(data, &probe)

	loaded := defaultConfig
	if preset, ok := ColorPresets[probe.Theme]; ok {
		loaded.Theme = probe.Theme
		loaded.BackgroundColor = preset.BackgroundColor
		loaded.BackgroundAlpha = preset.BackgroundAlpha
		loaded.TextColor = preset.TextColor
		loaded.SelectionColor = preset.SelectionColor
		loaded.FocusBorderColor = preset.FocusBorderColor
		loaded.EntryBackground = preset.EntryBackground
		loaded.DescColor = preset.DescColor
	}

	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return err
	}
	Current = loaded
	return nil
}

func Save() error {
	configDir := ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(&Current)
	if err != nil {
		return err
	}

	return AtomicWrite(configPath, data, 0644)
}

func AtomicWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := tmpFile.Name()
	if err := tmpFile.Chmod(mode); err != nil {
		tmpFile.Close()
		os.Remove(tmp)
		return err
	}
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmp)
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmp)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func SetupHotkey(edwinPath string) error {
	bindPath := HomeFile(".config", "mango", "bind.conf")
	bindLine := fmt.Sprintf("bind=%s,spawn,%s", Current.Hotkey, edwinPath)

	data, err := os.ReadFile(bindPath)
	if err != nil {
		os.MkdirAll(filepath.Dir(bindPath), 0755)
		return AtomicWrite(bindPath, []byte(bindLine+"\n"), 0644)
	}

	content := string(data)
	var newLines []string
	found := false

	for _, line := range splitLines(content) {
		if containsEdwinBind(line) {
			if !found {
				newLines = append(newLines, bindLine)
				found = true
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, bindLine)
	}

	return AtomicWrite(bindPath, []byte(joinLines(newLines)), 0644)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	if len(result) > 0 && result[len(result)-1] != '\n' {
		result += "\n"
	}
	return result
}

func containsEdwinBind(line string) bool {
	for i := 0; i < len(line); i++ {
		if i+5 <= len(line) && line[i:i+5] == "edwin" {
			return true
		}
	}
	return false
}

func GetCSS() string {
	c := Current
	return StaticThemeCSS + fmt.Sprintf(`
		@define-color edwin-bg rgba(%s, %.2f);
		@define-color edwin-text %s;
		@define-color edwin-selection rgba(%s, 0.3);
		@define-color edwin-focus-border rgba(%s, 1);
		@define-color edwin-entry-bg %s;
		@define-color edwin-desc-color %s;

		window {
			background: @edwin-bg;
			border-radius: %dpx;
		}
		#edwin-entry {
			font-size: %dpx;
			padding: %dpx;
			background: @edwin-entry-bg;
			border: 1px solid transparent;
			border-radius: %dpx;
			color: @edwin-text;
		}
		#edwin-entry:focus {
			border: 1px solid @edwin-focus-border;
		}
		#edwin-results {
			background: transparent;
		}
		#edwin-row {
			background: transparent;
			color: @edwin-text;
			border-radius: %dpx;
			padding: %dpx %dpx;
			outline: none;
			box-shadow: none;
		}
		#edwin-row:hover,
		#edwin-row:focus,
		#edwin-row:active {
			background: transparent;
			outline: none;
			box-shadow: none;
		}
		#edwin-row:selected,
		#edwin-row:selected:hover,
		#edwin-row:selected:focus {
			background: @edwin-selection;
			outline: none;
			box-shadow: none;
		}
		#edwin-title {
			color: @edwin-text;
			font-size: %dpx;
		}
		#edwin-desc {
			color: @edwin-desc-color;
			font-size: %dpx;
		}
		#edwin-shortcut-badge {
			color: @edwin-desc-color;
			font-size: 11px;
		}
	`, c.BackgroundColor, c.BackgroundAlpha, c.TextColor, c.SelectionColor,
		c.FocusBorderColor, c.EntryBackground, c.DescColor,
		c.BorderRadius, c.FontSize, SpacingLarge, RadiusMedium,
		RadiusSmall, SpacingTiny, SpacingMedium, FontSizeTitle, FontSizeSmall)
}
