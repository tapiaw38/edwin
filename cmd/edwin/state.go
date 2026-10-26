package main

import (
	"context"
	"sync"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/edwin/internal/apps"
	"github.com/tapiaw38/edwin/internal/modules"
)

var (
	listBox        *gtk.ListBox
	resultsScroll  *gtk.ScrolledWindow
	searchEntry    *gtk.Entry
	previewBox     *gtk.Box
	sidePanel      *gtk.Box
	previewToolbar *gtk.Box
	previewMeta    *gtk.Label
	previewImage   *gtk.Image
	previewLabel   *gtk.Label
)

var (
	allApps        []apps.App
	currentResults []modules.Result
	searchVersion  uint64
	searchCancel   context.CancelFunc
	searchCancelMu sync.Mutex
	debounceTimer  *time.Timer
	inActionMode   bool
)

var (
	resultsRenderVersion uint64
	resultsRenderedN     int
	resultsLoadingMore   bool
	isClearing           bool
)

var (
	previewVersion  uint64
	quickLookActive bool

	shiftHeld           bool
	shiftUsedAsModifier bool
	previewPage         int
	previewScale        int
)

var (
	spotifyView       *gtk.Box
	spotifyScroll     *gtk.ScrolledWindow
	sidePanelSequence uint64
	spotifyArtBig     *gtk.Image
	spotifyTitle      *gtk.Label
	spotifyArtist     *gtk.Label
	spotifyAlbum      *gtk.Label
	spotifyStatus     *gtk.Label
	spotifyList       *gtk.ListBox
	inSpotifyMode     bool
	playerMode        modules.PlayerKind
	playerInfoVersion uint64
)

var (
	iconCache   = make(map[string]*gdkpixbuf.Pixbuf)
	iconCacheMu sync.RWMutex
)

const (
	dragInfoURI uint = iota + 1
	dragInfoText
)
