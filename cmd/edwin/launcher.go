package main

import (
	_ "embed"
	"os"
	"time"

	"github.com/diamondburned/gotk4-layer-shell/pkg/gtklayershell"
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/gogpu/systray"
	"github.com/tapiaw38/edwin/internal/apps"
	"github.com/tapiaw38/edwin/internal/config"
)

const (
	searchDebounce = 50 * time.Millisecond

	resultsHeight = 288
	previewHeight = 232

	previewScaleDefault = 360
	previewScaleMin     = 180
	previewScaleMax     = 720
	previewScaleStep    = 60

	shortQueryLen = 1
	preloadApps   = 20
)

//go:embed assets/edwin-tray.png
var trayIcon []byte

func runLauncher() {
	activationRequests := make(chan struct{}, 1)
	cleanupInstance, alreadyRunning, err := claimSingleInstance(func() {
		select {
		case activationRequests <- struct{}{}:
		default:
		}
	})
	if alreadyRunning {
		if err != nil {
			os.Stderr.WriteString("Failed to activate running Edwin instance: " + err.Error() + "\n")
		}
		return
	}
	if err != nil {
		os.Stderr.WriteString("Failed to register Edwin instance: " + err.Error() + "\n")
		return
	}
	defer cleanupInstance()

	gtk.Init()
	if err := config.Load(); err != nil {
		os.Stderr.WriteString("Edwin config invalid: " + err.Error() + "\n")
	}

	allApps = apps.Load()
	initSearch()
	go preloadIcons()

	window := newLauncherWindow()
	mainBox := gtk.NewBox(gtk.OrientationVertical, 8)
	mainBox.SetMarginStart(12)
	mainBox.SetMarginEnd(12)
	mainBox.SetMarginTop(12)
	mainBox.SetMarginBottom(12)

	createSpotifyView()
	entry := buildSearchEntry()
	buildResultsList()
	buildPreviewPane()
	connectKeyBindings(window)

	window.ConnectDeleteEvent(func(_ *gdk.Event) bool {
		return window.HideOnDelete()
	})

	mainBox.PackStart(entry, false, false, 0)
	mainBox.PackStart(resultsScroll, false, false, 0)
	mainBox.PackStart(spotifyView, false, false, 0)
	mainBox.PackStart(previewBox, false, false, 0)
	window.Add(mainBox)

	loadCSS()

	window.ShowAll()
	spotifyView.Hide()
	resultsScroll.Hide()
	entry.GrabFocus()
	tray := systray.New()
	menu := systray.NewMenu()
	menu.Add("Quit", func() {
		glib.IdleAdd(func() { gtk.MainQuit() })
	})
	tray.SetIcon(trayIcon).SetTooltip("Edwin").OnClick(func() {
		glib.IdleAdd(func() {
			if window.IsVisible() {
				window.Hide()
			} else {
				window.Show()
				window.Present()
				entry.GrabFocus()
			}
		})
	}).SetMenu(menu).Show()
	go func() {
		for range activationRequests {
			glib.IdleAdd(func() {
				window.Show()
				window.Present()
				entry.GrabFocus()
			})
		}
	}()

	gtk.Main()
	tray.Remove()
}

func newLauncherWindow() *gtk.Window {
	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Edwin")
	window.SetDefaultSize(config.Current.Width, -1)
	window.SetSizeRequest(config.Current.Width, -1)
	window.SetDecorated(false)

	gtklayershell.InitForWindow(window)
	gtklayershell.SetLayer(window, gtklayershell.LayerShellLayerTop)
	gtklayershell.SetKeyboardMode(window, gtklayershell.LayerShellKeyboardModeExclusive)
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeTop, true)
	gtklayershell.SetMargin(window, gtklayershell.LayerShellEdgeTop, config.Current.MarginTop)
	return window
}

func buildSearchEntry() *gtk.Entry {
	entry := gtk.NewEntry()
	searchEntry = entry
	entry.SetPlaceholderText("Search apps, ;snippet, define word, > shell...")
	entry.SetName("edwin-entry")

	entry.Connect("changed", func() {
		query := entry.Text()
		inActionMode = false

		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		if len(query) <= shortQueryLen {
			updateResults(query)
			return
		}

		debounceTimer = time.AfterFunc(searchDebounce, func() {
			glib.IdleAdd(func() { updateResults(query) })
		})
	})
	return entry
}

func buildResultsList() {
	listBox = gtk.NewListBox()
	listBox.SetName("edwin-results")
	listBox.SetSelectionMode(gtk.SelectionSingle)

	resultsScroll = gtk.NewScrolledWindow(nil, nil)
	resultsScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	resultsScroll.SetSizeRequest(-1, resultsHeight)
	resultsScroll.SetNoShowAll(true)
	resultsScroll.Add(listBox)
	resultsScroll.VAdjustment().Connect("value-changed", maybeLoadMoreResults)

	listBox.Connect("row-selected", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		if quickLookActive {
			updatePreview(row)
		} else {
			hidePreview()
		}
	})
}

func buildPreviewPane() {
	previewBox = gtk.NewBox(gtk.OrientationVertical, 8)
	previewBox.SetName("edwin-preview")
	previewBox.SetSizeRequest(-1, previewHeight)
	previewBox.SetNoShowAll(true)

	previewBox.PackStart(buildPreviewToolbar(), false, false, 0)

	previewImage = gtk.NewImage()
	previewImage.SetName("edwin-preview-image")
	previewImage.SetSizeRequest(220, 180)
	previewImage.SetNoShowAll(true)

	previewLabel = gtk.NewLabel("")
	previewLabel.SetName("edwin-preview-label")
	previewLabel.SetXAlign(0)
	previewLabel.SetLineWrap(true)
	previewLabel.SetLineWrapMode(pango.WrapWordChar)
	previewLabel.SetMaxWidthChars(40)
	previewLabel.SetSizeRequest(300, -1)
	previewLabel.SetNoShowAll(true)

	previewBox.PackStart(previewImage, false, false, 0)
	previewBox.PackStart(previewLabel, false, false, 0)
}

func buildPreviewToolbar() *gtk.Box {
	previewToolbar = gtk.NewBox(gtk.OrientationHorizontal, 6)
	previewToolbar.SetName("edwin-preview-toolbar")
	previewToolbar.SetNoShowAll(true)

	buttons := []struct {
		label string
		click func()
	}{
		{"‹", func() { changePreviewPage(-1) }},
		{"›", func() { changePreviewPage(1) }},
		{"-", func() { changePreviewZoom(-previewScaleStep) }},
		{"+", func() { changePreviewZoom(previewScaleStep) }},
	}
	for _, b := range buttons {
		btn := gtk.NewButtonWithLabel(b.label)
		btn.Connect("clicked", b.click)
		previewToolbar.PackStart(btn, false, false, 0)
	}

	previewMeta = gtk.NewLabel("")
	previewMeta.SetXAlign(0)
	previewMeta.SetName("edwin-desc")
	previewToolbar.PackStart(previewMeta, true, true, 0)
	return previewToolbar
}

func connectKeyBindings(window *gtk.Window) {
	window.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		switch event.AsKey().Keyval() {
		case gdk.KEY_Escape:
			if inActionMode {
				inActionMode = false
				updateResults(searchEntry.Text())
				return true
			}
			window.Hide()
			return true
		case gdk.KEY_Down:
			selectNext()
			return true
		case gdk.KEY_Up:
			selectPrev()
			return true
		case gdk.KEY_Return:
			executeSelected()
			return true
		case gdk.KEY_Tab:
			showSelectedFileActions()
			return true
		case gdk.KEY_Shift_L, gdk.KEY_Shift_R:
			toggleQuickLook()
			return true
		case gdk.KEY_Page_Down, gdk.KEY_Right:
			return quickLook(func() { changePreviewPage(1) })
		case gdk.KEY_Page_Up, gdk.KEY_Left:
			return quickLook(func() { changePreviewPage(-1) })
		case gdk.KEY_plus, gdk.KEY_KP_Add, gdk.KEY_equal:
			return quickLook(func() { changePreviewZoom(previewScaleStep) })
		case gdk.KEY_minus, gdk.KEY_KP_Subtract:
			return quickLook(func() { changePreviewZoom(-previewScaleStep) })
		}
		return false
	})
}

func quickLook(fn func()) bool {
	if !quickLookActive {
		return false
	}
	fn()
	return true
}

func toggleQuickLook() {
	quickLookActive = !quickLookActive
	if !quickLookActive {
		hidePreview()
		return
	}
	previewPage = 1
	if previewScale == 0 {
		previewScale = previewScaleDefault
	}
	updatePreview(listBox.SelectedRow())
}

func preloadIcons() {
	theme := gtk.IconThemeGetDefault()

	fileIcons := []string{
		"folder", "application-pdf", "x-office-document",
		"x-office-spreadsheet", "image-x-generic", "audio-x-generic",
		"video-x-generic", "text-x-script", "text-x-generic",
	}
	for _, name := range fileIcons {
		cacheIcon(theme, name, config.Current.IconSize)
	}

	for i, app := range allApps {
		if i >= preloadApps {
			break
		}
		if isThemeIconName(app.Icon) {
			cacheIcon(theme, app.Icon, 24)
		}
	}
}

func isThemeIconName(icon string) bool {
	return icon != "" && icon[0] != '/'
}

func cacheIcon(theme *gtk.IconTheme, name string, size int) {
	pb, err := theme.LoadIcon(name, size, gtk.IconLookupForceSize)
	if err != nil {
		return
	}
	iconCacheMu.Lock()
	defer iconCacheMu.Unlock()
	iconCache[name] = pb
}
