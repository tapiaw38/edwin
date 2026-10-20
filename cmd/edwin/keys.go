package main

import (
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/edwin/internal/config"
	"github.com/tapiaw38/edwin/internal/modules"
)

const shortcutRelevantModifiers = gdk.ControlMask | gdk.Mod1Mask | gdk.SuperMask | gdk.ShiftMask

func shouldToggleQuickLookOnShift(state gdk.ModifierType) bool {
	return state&(gdk.ControlMask|gdk.Mod1Mask|gdk.SuperMask) == 0
}

func shortcutModifier(name string) (gdk.ModifierType, string) {
	switch name {
	case "super":
		return gdk.SuperMask, "S"
	case "ctrl":
		return gdk.ControlMask, "C"
	case "alt":
		return gdk.Mod1Mask, "A"
	case "ctrl+shift":
		return gdk.ControlMask | gdk.ShiftMask, "C⇧"
	case "alt+shift":
		return gdk.Mod1Mask | gdk.ShiftMask, "A⇧"
	case "super+shift":
		return gdk.SuperMask | gdk.ShiftMask, "S⇧"
	default:
		return gdk.ControlMask | gdk.ShiftMask, "C⇧"
	}
}

func shortcutBadgeText(modifier string, index int) string {
	_, prefix := shortcutModifier(modifier)
	return prefix + "-" + strconv.Itoa(index+1)
}

func shortcutNumber(keyval uint, hardwareKeycode uint16) (int, bool) {
	if hardwareKeycode >= 10 && hardwareKeycode <= 18 {
		return int(hardwareKeycode - 10), true
	}
	switch keyval {
	case gdk.KEY_1, gdk.KEY_exclam:
		return 0, true
	case gdk.KEY_2, gdk.KEY_at, gdk.KEY_quotedbl:
		return 1, true
	case gdk.KEY_3, gdk.KEY_numbersign, gdk.KEY_section, gdk.KEY_periodcentered:
		return 2, true
	case gdk.KEY_4, gdk.KEY_dollar:
		return 3, true
	case gdk.KEY_5, gdk.KEY_percent:
		return 4, true
	case gdk.KEY_6, gdk.KEY_asciicircum, gdk.KEY_ampersand:
		return 5, true
	case gdk.KEY_7, gdk.KEY_slash:
		return 6, true
	case gdk.KEY_8, gdk.KEY_asterisk, gdk.KEY_parenleft:
		return 7, true
	case gdk.KEY_9, gdk.KEY_parenright:
		return 8, true
	case gdk.KEY_KP_1:
		return 0, true
	case gdk.KEY_KP_2:
		return 1, true
	case gdk.KEY_KP_3:
		return 2, true
	case gdk.KEY_KP_4:
		return 3, true
	case gdk.KEY_KP_5:
		return 4, true
	case gdk.KEY_KP_6:
		return 5, true
	case gdk.KEY_KP_7:
		return 6, true
	case gdk.KEY_KP_8:
		return 7, true
	case gdk.KEY_KP_9:
		return 8, true
	default:
		return 0, false
	}
}

func executeShortcut(window *gtk.Window, keyval uint, hardwareKeycode uint16, state gdk.ModifierType) bool {
	if inSpotifyMode || listBox == nil {
		return false
	}
	modifier, _ := shortcutModifier(config.Current.ShortcutModifier)
	if state&shortcutRelevantModifiers != modifier {
		return false
	}
	index, ok := shortcutNumber(keyval, hardwareKeycode)
	if !ok || index >= len(currentResults) {
		return false
	}
	for resultsRenderedN <= index {
		if !renderNextResultChunk() {
			return false
		}
	}
	row := listBox.RowAtIndex(index)
	if row == nil {
		return false
	}
	selectRow(listBox, row)
	executeSelected(window)
	return true
}

func activeList() *gtk.ListBox {
	if inSpotifyMode && spotifyList != nil {
		return spotifyList
	}
	return listBox
}

func selectNext() {
	list := activeList()
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil {
		selectRow(list, list.RowAtIndex(0))
		return
	}

	idx := selected.Index()
	next := list.RowAtIndex(idx + 1)
	if next == nil && list == listBox && resultsRenderedN < len(currentResults) {
		renderNextResultChunk()
		next = list.RowAtIndex(idx + 1)
	}
	selectRow(list, next)
}

func selectPrev() {
	list := activeList()
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil || selected.Index() <= 0 {
		return
	}
	selectRow(list, list.RowAtIndex(selected.Index()-1))
}

func selectRow(list *gtk.ListBox, row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	list.SelectRow(row)
	scrollToRow(row)
}

func scrollToRow(row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	scroll := resultsScroll
	if inSpotifyMode {
		scroll = spotifyScroll
	}
	if scroll == nil {
		return
	}
	adj := scroll.VAdjustment()
	if adj == nil {
		return
	}

	alloc := row.Allocation()
	rowTop := float64(alloc.Y())
	rowBottom := rowTop + float64(alloc.Height())
	viewTop := adj.Value()
	viewBottom := viewTop + adj.PageSize()

	switch {
	case rowBottom > viewBottom:
		adj.SetValue(rowBottom - adj.PageSize())
	case rowTop < viewTop:
		adj.SetValue(rowTop)
	}
}

func executeSelected(window *gtk.Window) {
	if inSpotifyMode && spotifyList != nil {
		executePlayerControl()
		return
	}

	selected := listBox.SelectedRow()
	r, ok := selectedResult(selected)
	if !ok {
		return
	}

	if r.NavigateQuery != "" {
		searchEntry.SetText(r.NavigateQuery)
		searchEntry.SetPosition(-1)
		return
	}
	if r.Confirm && !confirmAction(r) {
		return
	}
	executeActionSpec(r.ActionSpec)
	if !r.KeepOpen {
		hideLauncher(window)
		return
	}
	if inActionMode {
		inActionMode = false
		updateResults(searchEntry.Text())
	}
}

func executePlayerControl() {
	selected := spotifyList.SelectedRow()
	if selected == nil || selected.Index() < 0 {
		return
	}
	controls := modules.PlayerControls(playerMode)
	if selected.Index() >= len(controls) {
		return
	}
	runPlayerResultAction(controls[selected.Index()])
}

func confirmAction(r modules.Result) bool {
	dialog := gtk.NewMessageDialog(mainWindow, gtk.DialogModal, gtk.MessageWarning, gtk.ButtonsOKCancel)
	dialog.SetMarkup("<b>" + r.Title + "</b>\n" + r.Desc)
	dialog.ShowAll()
	response := dialog.Run()
	dialog.Destroy()
	return gtk.ResponseType(response) == gtk.ResponseOK
}
