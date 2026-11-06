package config

import (
	"strings"
	"testing"
)

func TestEdwinThemeIsDefault(t *testing.T) {
	if defaultConfig.Theme != "edwin" {
		t.Fatalf("default theme = %q, want edwin", defaultConfig.Theme)
	}
	if defaultConfig.Position != "top" {
		t.Fatalf("default position = %q, want top", defaultConfig.Position)
	}
	if defaultConfig.ShortcutModifier != "ctrl+shift" {
		t.Fatalf("default shortcut modifier = %q, want ctrl+shift", defaultConfig.ShortcutModifier)
	}
	if got := ColorPresets["edwin"].FocusBorderColor; got != "245, 215, 140" {
		t.Fatalf("edwin focus border = %q, want 245, 215, 140", got)
	}
}

func TestGetCSSUsesFocusBorderColor(t *testing.T) {
	original := Current
	Current = defaultConfig
	t.Cleanup(func() { Current = original })

	css := GetCSS()
	if !strings.Contains(css, "@define-color edwin-focus-border rgba(245, 215, 140, 1);") {
		t.Fatal("CSS does not define the edwin focus border color")
	}
	if !strings.Contains(css, "#edwin-entry:focus {") || !strings.Contains(css, "border: 1px solid @edwin-focus-border;") {
		t.Fatal("CSS does not apply the focus border to the input")
	}
}

func TestGetCSSUsesConfiguredBorder(t *testing.T) {
	previous := Current
	defer func() { Current = previous }()

	Current = defaultConfig
	Current.BorderWidth = 3
	Current.BorderColor = "#ff0000"
	css := GetCSS()
	if !strings.Contains(css, "border: 3px solid @edwin-border;") {
		t.Fatal("CSS does not apply the configured border width")
	}
	if !strings.Contains(css, "@define-color edwin-border #ff0000;") {
		t.Fatal("CSS does not define the configured border color")
	}

	Current.BorderColor = ""
	if css := GetCSS(); !strings.Contains(css, "@define-color edwin-border rgba(245, 215, 140, 1);") {
		t.Fatal("CSS does not fall back to the focus border color")
	}
}

func TestGetCSSUsesConfiguredInputBorder(t *testing.T) {
	previous := Current
	defer func() { Current = previous }()

	Current = defaultConfig
	css := GetCSS()
	if !strings.Contains(css, "@define-color edwin-input-border transparent;") {
		t.Fatal("unset input border color should be transparent")
	}
	if !strings.Contains(css, "border: 1px solid @edwin-input-border;") {
		t.Fatal("CSS does not apply the default input border width")
	}

	Current.InputBorderWidth = 2
	Current.InputBorderColor = "208, 193, 180"
	css = GetCSS()
	if !strings.Contains(css, "@define-color edwin-input-border rgba(208, 193, 180, 1);") {
		t.Fatal("CSS does not define the configured input border color")
	}
	if !strings.Contains(css, "border: 2px solid @edwin-input-border;") {
		t.Fatal("CSS does not apply the configured input border width")
	}
	if !strings.Contains(css, "border: 2px solid @edwin-focus-border;") {
		t.Fatal("focus border should use the same width")
	}

	Current.InputBorderWidth = 0
	if css := GetCSS(); !strings.Contains(css, "border: 1px solid @edwin-focus-border;") {
		t.Fatal("focus border should keep 1px when the input border is disabled")
	}
}

func TestGetCSSUsesConfiguredInnerRadius(t *testing.T) {
	previous := Current
	defer func() { Current = previous }()

	Current = defaultConfig
	Current.BorderRadius = 0
	css := GetCSS()
	if strings.Count(css, "border-radius: 0px;") != 3 {
		t.Fatalf("window, input and row radius should all inherit border_radius:\n%s", css)
	}

	override := 14
	Current.InputBorderRadius = &override
	css = GetCSS()
	if !strings.Contains(css, "border-radius: 14px;") {
		t.Fatal("explicit input_border_radius should override the inherited value")
	}
	if strings.Count(css, "border-radius: 0px;") != 2 {
		t.Fatal("window and row radius should still inherit border_radius")
	}

	rowRadius := 6
	Current.RowBorderRadius = &rowRadius
	css = GetCSS()
	if !strings.Contains(css, "border-radius: 6px;") {
		t.Fatal("explicit row_border_radius should override the inherited value")
	}
	if strings.Count(css, "border-radius: 0px;") != 1 {
		t.Fatal("window radius should still inherit border_radius")
	}
}
