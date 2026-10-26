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
	searchDebounce        = 50 * time.Millisecond
	showAnimationSteps    = 12
	showAnimationInterval = 15

	resultsHeight  = 288
	previewHeight  = 232
	sidePanelWidth = 300
	mainBoxSpacing = 8
	mainBoxMargin  = 12

	previewScaleDefault = 360
	previewScaleMin     = 180
	previewScaleMax     = 720
	previewScaleStep    = 60

	shortQueryLen = 1
	preloadApps   = 20
)

//go:embed assets/edwin-tray.png
var trayIcon []byte
var mainWindow *gtk.Window

type positionState struct {
	sequence        uint64
	monitorHeight   int
	collapsedHeight int
	centerInset     int
	edge            gtklayershell.Edge
	targetMargin    int
	animating       bool
}

var launcherPosition positionState

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
	mainWindow = window
	mainBox := gtk.NewBox(gtk.OrientationHorizontal, mainBoxSpacing)
	mainBox.SetMarginStart(mainBoxMargin)
	mainBox.SetMarginEnd(mainBoxMargin)
	mainBox.SetMarginTop(mainBoxMargin)
	mainBox.SetMarginBottom(mainBoxMargin)

	createSpotifyView()
	entry := buildSearchEntry()
	buildResultsList()
	resultsScroll.SetSizeRequest(config.Current.Width-2*mainBoxMargin, resultsHeight)
	buildPreviewPane()
	buildSidePanel()
	spotifyScroll.SetSizeRequest(config.Current.Width-2*mainBoxMargin, resultsHeight)
	connectKeyBindings(window)

	window.ConnectDeleteEvent(func(_ *gdk.Event) bool {
		hideLauncher(window)
		return true
	})

	leftColumn := gtk.NewBox(gtk.OrientationVertical, mainBoxSpacing)
	leftColumn.SetSizeRequest(config.Current.Width-2*mainBoxMargin, -1)
	leftColumn.PackStart(entry, false, false, 0)
	leftColumn.PackStart(resultsScroll, true, true, 0)
	leftColumn.PackStart(spotifyScroll, true, true, 0)
	mainBox.PackStart(leftColumn, false, false, 0)
	mainBox.PackEnd(sidePanel, false, false, 0)
	window.Add(mainBox)

	loadCSS()
	window.SetOpacity(0)
	window.ShowAll()
	hideLauncher(window)
	entry.GrabFocus()
	animateShow(window)
	tray := systray.New()
	menu := systray.NewMenu()
	menu.Add("Quit", func() {
		glib.IdleAdd(func() { gtk.MainQuit() })
	})
	tray.SetIcon(trayIcon).SetTooltip("Edwin").OnClick(func() {
		glib.IdleAdd(func() {
			if window.IsVisible() {
				hideLauncher(window)
			} else {
				animateShow(window)
				entry.GrabFocus()
			}
		})
	}).SetMenu(menu).Show()
	go func() {
		for range activationRequests {
			glib.IdleAdd(func() {
				animateShow(window)
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
	return window
}

func animateShow(window *gtk.Window) {
	if window.IsVisible() {
		window.Present()
		return
	}
	launcherPosition.sequence++
	sequence := launcherPosition.sequence
	position := config.Current.Position
	edge := gtklayershell.LayerShellEdgeTop
	if position == "bottom" {
		edge = gtklayershell.LayerShellEdgeBottom
	} else if position != "center" {
		position = "top"
	}
	display := gdk.DisplayGetDefault()
	estimatedMonitor := fallbackMonitor(display)
	estimatedHeight, estimatedWidth, estimatedInset := monitorDimensions(estimatedMonitor)
	collapsedHeight := launcherPosition.collapsedHeight
	if collapsedHeight <= 0 {
		collapsedHeight = estimatedCollapsedHeight()
	}
	estimatedLeftMargin := centeredLeftMargin(estimatedWidth, config.Current.Width)
	estimatedVerticalMargin := 0
	if position == "center" {
		estimatedVerticalMargin = centeredInputTopMargin(estimatedHeight, collapsedHeight, estimatedInset)
	}
	launcherPosition.edge = edge
	launcherPosition.targetMargin = estimatedVerticalMargin
	launcherPosition.monitorHeight = estimatedHeight
	launcherPosition.centerInset = estimatedInset
	launcherPosition.animating = true
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeTop, false)
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeBottom, false)
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeLeft, true)
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeRight, false)
	gtklayershell.SetAnchor(window, edge, true)
	applyPositionMargin(window, edge, estimatedVerticalMargin)
	gtklayershell.SetMargin(window, gtklayershell.LayerShellEdgeLeft, estimatedLeftMargin)
	window.SetOpacity(0)
	window.Show()
	window.Present()
	glib.IdleAdd(func() {
		if sequence != launcherPosition.sequence {
			return
		}
		monitor := monitorForWindow(display, window)
		monitorHeight, monitorWidth, centerInset := monitorDimensions(monitor)
		windowHeight := window.AllocatedHeight()
		if windowHeight <= 0 {
			_, windowHeight = window.Size()
		}
		if windowHeight > 0 && launcherIsCollapsed() {
			launcherPosition.collapsedHeight = windowHeight
		}
		launcherPosition.monitorHeight = monitorHeight
		launcherPosition.centerInset = centerInset
		actualLeftMargin := centeredLeftMargin(monitorWidth, config.Current.Width)
		if actualLeftMargin != estimatedLeftMargin {
			gtklayershell.SetMargin(window, gtklayershell.LayerShellEdgeLeft, actualLeftMargin)
		}
		if position == "center" {
			actualVerticalMargin := centeredInputTopMargin(monitorHeight, windowHeight, centerInset)
			if actualVerticalMargin != estimatedVerticalMargin {
				applyPositionMargin(window, edge, actualVerticalMargin)
			}
			launcherPosition.targetMargin = actualVerticalMargin
		} else {
			updatePositionForHeight(window, windowHeight, false)
		}
		glib.TimeoutAdd(24, func() bool {
			if sequence != launcherPosition.sequence || !window.IsVisible() {
				return false
			}
			step := 0
			glib.TimeoutAdd(showAnimationInterval, func() bool {
				if sequence != launcherPosition.sequence || !window.IsVisible() {
					return false
				}
				step++
				progress := float64(step) / showAnimationSteps
				window.SetOpacity(progress)
				if position != "center" {
					margin := int(float64(launcherPosition.targetMargin) * progress)
					applyPositionMargin(window, edge, margin)
				}
				if step >= showAnimationSteps {
					launcherPosition.animating = false
					window.SetOpacity(1)
					if position != "center" {
						applyPositionMargin(window, edge, launcherPosition.targetMargin)
					}
					return false
				}
				return true
			})
			return false
		})
	})
}

func estimatedCollapsedHeight() int {
	height := config.Current.FontSize + 2*config.SpacingLarge + 2*mainBoxMargin + 2
	if height < 1 {
		return 1
	}
	return height
}

func launcherIsCollapsed() bool {
	return !resultsScroll.Visible() && !spotifyScroll.Visible() && !previewBox.Visible() && !spotifyView.Visible()
}

func monitorForWindow(display *gdk.Display, window *gtk.Window) *gdk.Monitor {
	if display == nil {
		return nil
	}
	if nativeWindow := window.Window(); nativeWindow != nil {
		if monitor := display.MonitorAtWindow(nativeWindow); monitor != nil {
			return monitor
		}
	}
	return fallbackMonitor(display)
}

func fallbackMonitor(display *gdk.Display) *gdk.Monitor {
	if display == nil {
		return nil
	}
	if monitor := display.PrimaryMonitor(); monitor != nil {
		return monitor
	}
	if display.NMonitors() > 0 {
		return display.Monitor(0)
	}
	return nil
}

func monitorDimensions(monitor *gdk.Monitor) (height, width, centerInset int) {
	if monitor != nil {
		if geometry := monitor.Geometry(); geometry != nil {
			geometryHeight := geometry.Height()
			height = geometryHeight
			width = geometry.Width()
			if workarea := monitor.Workarea(); workarea != nil && workarea.Height() > 0 && workarea.Height() < geometryHeight {
				height = workarea.Height()
			} else {
				centerInset = 32
			}
		}
	}
	if height <= 0 {
		if screen := gdk.ScreenGetDefault(); screen != nil {
			height = screen.Height()
			width = screen.Width()
			centerInset = 32
		}
	}
	if height <= 0 {
		height = 1080
		centerInset = 32
	}
	if width <= 0 {
		width = config.Current.Width
	}
	return
}

func centeredLeftMargin(monitorWidth, collapsedWidth int) int {
	margin := (monitorWidth - collapsedWidth) / 2
	if margin < 0 {
		return 0
	}
	return margin
}

func centeredInputTopMargin(monitorHeight, collapsedHeight, centerInset int) int {
	margin := monitorHeight/2 - centerInset - 2*collapsedHeight
	if margin < 0 {
		return 0
	}
	return margin
}

func updatePositionForHeight(window *gtk.Window, windowHeight int, apply bool) int {
	margin := (launcherPosition.monitorHeight-windowHeight)/2 - launcherPosition.centerInset
	if margin < 0 {
		margin = 0
	}
	launcherPosition.targetMargin = margin
	if apply {
		applyPositionMargin(window, launcherPosition.edge, margin)
	}
	return margin
}

func applyPositionMargin(window *gtk.Window, edge gtklayershell.Edge, margin int) {
	gtklayershell.SetMargin(window, edge, margin)
}

func hideLauncher(window *gtk.Window) {
	inActionMode = false
	quickLookActive = false
	if searchEntry.Text() == "" {
		updateResults("")
	} else {
		searchEntry.SetText("")
	}
	resetShiftTap()
	resultsScroll.Hide()
	spotifyView.Hide()
	previewBox.Hide()
	window.Hide()
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
		updateShortcutBadges(row)
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

func buildSidePanel() {
	sidePanel = gtk.NewBox(gtk.OrientationVertical, 0)
	sidePanel.SetName("edwin-side-panel")
	sidePanel.SetSizeRequest(sidePanelWidth, -1)
	sidePanel.SetNoShowAll(true)
	sidePanel.PackStart(previewBox, true, true, 0)
	sidePanel.PackStart(spotifyView, true, true, 0)

	previewBox.ConnectShow(showSidePanel)
	previewBox.ConnectHide(syncSidePanelVisibility)
	spotifyView.ConnectShow(showSidePanel)
	spotifyView.ConnectHide(syncSidePanelVisibility)
}

func showSidePanel() {
	sidePanelSequence++
	sequence := sidePanelSequence
	sidePanel.SetOpacity(0)
	sidePanel.Show()
	step := 0
	glib.TimeoutAdd(showAnimationInterval, func() bool {
		if sequence != sidePanelSequence || !sidePanel.Visible() {
			return false
		}
		step++
		sidePanel.SetOpacity(float64(step) / showAnimationSteps)
		if step >= showAnimationSteps {
			sidePanel.SetOpacity(1)
			return false
		}
		return true
	})
}

func syncSidePanelVisibility() {
	if previewBox.Visible() || spotifyView.Visible() {
		return
	}
	sidePanelSequence++
	sidePanel.SetOpacity(1)
	sidePanel.Hide()
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
	window.AddEvents(int(gdk.KeyPressMask | gdk.KeyReleaseMask))
	window.Connect("key-release-event", func(_ *gtk.Window, event *gdk.Event) bool {
		keyEvent := event.AsKey()
		if !isShiftKey(keyEvent.Keyval()) {
			return false
		}
		tapped := shiftHeld && !shiftUsedAsModifier
		resetShiftTap()
		if !tapped || !shouldToggleQuickLookOnShift(keyEvent.State()) {
			return false
		}
		toggleQuickLook()
		return true
	})
	window.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		keyEvent := event.AsKey()
		if isShiftKey(keyEvent.Keyval()) {
			if !shiftHeld {
				shiftHeld = true
				shiftUsedAsModifier = !shouldToggleQuickLookOnShift(keyEvent.State())
			}
			return false
		}
		shiftUsedAsModifier = true
		if executeShortcut(window, keyEvent.Keyval(), keyEvent.HardwareKeycode(), keyEvent.State()) {
			return true
		}
		switch keyEvent.Keyval() {
		case gdk.KEY_Escape:
			if inActionMode {
				inActionMode = false
				updateResults(searchEntry.Text())
				return true
			}
			hideLauncher(window)
			return true
		case gdk.KEY_Down:
			if showRecentApps() {
				return true
			}
			selectNext()
			return true
		case gdk.KEY_Up:
			selectPrev()
			return true
		case gdk.KEY_Return:
			executeSelected(window)
			return true
		case gdk.KEY_Tab:
			showSelectedFileActions()
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
