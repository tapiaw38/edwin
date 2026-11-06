package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Width             int     `yaml:"width"`
	MaxResults        int     `yaml:"max_results"`
	Theme             string  `yaml:"theme"`
	BackgroundColor   string  `yaml:"background_color"`
	BackgroundAlpha   float64 `yaml:"background_alpha"`
	BorderRadius      int     `yaml:"border_radius"`
	BorderWidth       int     `yaml:"border_width"`
	BorderColor       string  `yaml:"border_color"`
	InputBorderWidth  int     `yaml:"input_border_width"`
	InputBorderColor  string  `yaml:"input_border_color"`
	InputBorderRadius *int    `yaml:"input_border_radius,omitempty"`
	RowBorderRadius   *int    `yaml:"row_border_radius,omitempty"`
	FontSize          int     `yaml:"font_size"`
	FontFamily        string  `yaml:"font_family"`
	TextColor         string  `yaml:"text_color"`
	SelectionColor    string  `yaml:"selection_color"`
	FocusBorderColor  string  `yaml:"focus_border_color"`
	EntryBackground   string  `yaml:"entry_background"`
	DescColor         string  `yaml:"desc_color"`

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
	BackgroundColor:  "28, 25, 22",
	BackgroundAlpha:  0.97,
	BorderRadius:     12,
	BorderWidth:      0,
	InputBorderWidth: 1,
	FontSize:         18,
	FontFamily:       `"Inter", "Noto Sans", sans-serif`,
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

func colorValue(value string) string {
	if !strings.Contains(value, ",") || value[0] < '0' || value[0] > '9' {
		return value
	}
	return "rgba(" + value + ", 1)"
}

func borderColorValue(c Config) string {
	if c.BorderColor != "" {
		return colorValue(c.BorderColor)
	}
	return colorValue(c.FocusBorderColor)
}

func radiusOrInherit(value *int, inherited int) int {
	if value == nil {
		return inherited
	}
	return *value
}

func inputBorderColorValue(c Config) string {
	if c.InputBorderColor == "" {
		return "transparent"
	}
	return colorValue(c.InputBorderColor)
}

func inputFocusBorderWidth(c Config) int {
	if c.InputBorderWidth < 1 {
		return 1
	}
	return c.InputBorderWidth
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
		@define-color edwin-border %s;
		@define-color edwin-input-border %s;
		@define-color edwin-query-icon rgba(%s, 0.72);

		window {
			font-family: %s;
			background: @edwin-bg;
			border-radius: %dpx;
			border: %dpx solid @edwin-border;
		}
		#edwin-entry {
			font-size: %dpx;
			font-weight: 500;
			padding: %dpx %dpx;
			background: @edwin-entry-bg;
			border: %dpx solid @edwin-input-border;
			border-radius: %dpx;
			color: @edwin-text;
		}
		#edwin-entry:focus {
			border: %dpx solid @edwin-focus-border;
			box-shadow: 0 0 0 2px alpha(@edwin-focus-border, 0.18);
		}
		#edwin-entry image {
			color: @edwin-query-icon;
			opacity: 0.72;
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
			background: alpha(@edwin-focus-border, 0.08);
			outline: none;
			box-shadow: none;
		}
		#edwin-row:selected,
		#edwin-row:selected:hover,
		#edwin-row:selected:focus {
			background: alpha(@edwin-selection, 0.42);
			outline: none;
		}
		#edwin-row:selected #edwin-title {
			font-weight: 700;
		}
		#edwin-row:selected #edwin-desc,
		#edwin-row:selected #edwin-shortcut-badge {
			color: @edwin-text;
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
			font-size: 10px;
			font-family: monospace;
			font-weight: 600;
			background: alpha(@edwin-focus-border, 0.12);
			border: 0;
			border-radius: 999px;
			min-height: 20px;
			padding: 0 10px;
		}
		#edwin-row:selected #edwin-shortcut-badge {
			background: alpha(@edwin-focus-border, 0.2);
		}
	`, c.BackgroundColor, c.BackgroundAlpha, c.TextColor, c.SelectionColor,
		c.FocusBorderColor, c.EntryBackground, c.DescColor, borderColorValue(c),
		inputBorderColorValue(c), c.BackgroundColor, c.FontFamily, c.BorderRadius, c.BorderWidth, c.FontSize,
		SpacingLarge, SpacingMedium, c.InputBorderWidth, radiusOrInherit(c.InputBorderRadius, c.BorderRadius),
		inputFocusBorderWidth(c), radiusOrInherit(c.RowBorderRadius, c.BorderRadius), SpacingTiny, SpacingMedium, FontSizeTitle, FontSizeSmall)
}
