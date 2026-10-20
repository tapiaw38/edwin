package main

import (
	"testing"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
)

func TestLookupCommand(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		wantOK   bool
		wantArgs []string
	}{
		{"no args runs the launcher", nil, false, nil},
		{"unknown flag runs the launcher", []string{"--nope"}, false, nil},
		{"flag with no argument", []string{"--stats-window"}, true, []string{}},
		{"flag with its argument", []string{"--large-type", "hello"}, true, []string{"hello"}},
		{"missing argument", []string{"--large-type"}, false, nil},
		{"file op needs three", []string{"--file-op-window", "copy", "a"}, false, nil},
		{"file op with three", []string{"--file-op-window", "copy", "a", "b"}, true, []string{"copy", "a", "b"}},
		{"email with none", []string{"--email-window"}, true, []string{}},
		{"email with some", []string{"--email-window", "me@example.com"}, true, []string{"me@example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, args, ok := lookupCommand(tt.argv)
			if ok != tt.wantOK {
				t.Fatalf("lookupCommand(%q) ok = %v, want %v", tt.argv, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("lookupCommand(%q) args = %q, want %q", tt.argv, args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestArgAt(t *testing.T) {
	args := []string{"a", "b"}
	for i, want := range []string{"a", "b", "", ""} {
		if got := argAt(args, i); got != want {
			t.Errorf("argAt(%q, %d) = %q, want %q", args, i, got, want)
		}
	}
}

func TestCenteredInputTopMargin(t *testing.T) {
	if got := centeredInputTopMargin(1080, 82, 32); got != 344 {
		t.Errorf("centeredInputTopMargin(1080, 82, 32) = %d, want 344", got)
	}
}

func TestCenteredLeftMargin(t *testing.T) {
	if got := centeredLeftMargin(1920, 600); got != 660 {
		t.Fatalf("centeredLeftMargin(1920, 600) = %d, want 660", got)
	}
	if got := centeredLeftMargin(500, 600); got != 0 {
		t.Fatalf("centeredLeftMargin(500, 600) = %d, want 0", got)
	}
}

func TestShortcutNumber(t *testing.T) {
	tests := []struct {
		keyval uint
		code   uint16
		want   int
		ok     bool
	}{
		{gdk.KEY_exclam, 10, 0, true},
		{gdk.KEY_periodcentered, 0, 2, true},
		{gdk.KEY_9, 0, 8, true},
		{gdk.KEY_parenright, 0, 8, true},
		{gdk.KEY_KP_5, 0, 4, true},
		{gdk.KEY_a, 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := shortcutNumber(tt.keyval, tt.code)
		if got != tt.want || ok != tt.ok {
			t.Errorf("shortcutNumber(%d, %d) = (%d, %t), want (%d, %t)", tt.keyval, tt.code, got, ok, tt.want, tt.ok)
		}
	}
}

func TestShortcutModifier(t *testing.T) {
	tests := []struct {
		name string
		mask gdk.ModifierType
		text string
	}{
		{"super", gdk.SuperMask, "S"},
		{"ctrl", gdk.ControlMask, "C"},
		{"alt", gdk.Mod1Mask, "A"},
		{"ctrl+shift", gdk.ControlMask | gdk.ShiftMask, "C⇧"},
		{"alt+shift", gdk.Mod1Mask | gdk.ShiftMask, "A⇧"},
		{"super+shift", gdk.SuperMask | gdk.ShiftMask, "S⇧"},
		{"invalid", gdk.ControlMask | gdk.ShiftMask, "C⇧"},
	}
	for _, tt := range tests {
		mask, text := shortcutModifier(tt.name)
		if mask != tt.mask || text != tt.text {
			t.Errorf("shortcutModifier(%q) = (%v, %q), want (%v, %q)", tt.name, mask, text, tt.mask, tt.text)
		}
	}
	if got := shortcutBadgeText("alt+shift", 8); got != "A⇧-9" {
		t.Errorf("shortcutBadgeText(alt+shift, 8) = %q, want A⇧-9", got)
	}
}

func TestShiftModifierDoesNotToggleQuickLook(t *testing.T) {
	for _, state := range []gdk.ModifierType{
		gdk.ControlMask,
		gdk.Mod1Mask,
		gdk.SuperMask,
		gdk.ControlMask | gdk.ShiftMask,
		gdk.Mod1Mask | gdk.ShiftMask,
		gdk.SuperMask | gdk.ShiftMask,
	} {
		if shouldToggleQuickLookOnShift(state) {
			t.Errorf("shouldToggleQuickLookOnShift(%v) = true", state)
		}
	}
	if !shouldToggleQuickLookOnShift(gdk.ShiftMask) {
		t.Error("plain Shift should toggle Quick Look")
	}
}
