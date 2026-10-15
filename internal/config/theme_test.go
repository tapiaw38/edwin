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
