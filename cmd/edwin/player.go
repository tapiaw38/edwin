package main

import (
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/edwin/internal/modules"
)

const (
	playerArtBigSize = 180

	playerRefreshDelay = 300 * time.Millisecond
)

func createSpotifyView() {
	spotifyView = gtk.NewBox(gtk.OrientationVertical, 8)
	spotifyView.SetName("spotify-view")
	spotifyArtBig = gtk.NewImage()
	spotifyArtBig.SetSizeRequest(playerArtBigSize, playerArtBigSize)
	spotifyArtBig.SetHAlign(gtk.AlignCenter)
	spotifyView.PackStart(spotifyArtBig, false, false, 0)

	spotifyTitle = gtk.NewLabel("")
	spotifyTitle.SetName("spotify-title")
	spotifyTitle.SetXAlign(0)
	spotifyTitle.SetEllipsize(pango.EllipsizeEnd)

	spotifyArtist = gtk.NewLabel("")
	spotifyArtist.SetName("spotify-artist")
	spotifyArtist.SetXAlign(0)

	spotifyAlbum = gtk.NewLabel("")
	spotifyAlbum.SetName("spotify-album")
	spotifyAlbum.SetXAlign(0)

	spotifyStatus = gtk.NewLabel("")
	spotifyStatus.SetName("spotify-status")
	spotifyStatus.SetXAlign(0)

	spotifyView.PackStart(spotifyTitle, false, false, 0)
	spotifyView.PackStart(spotifyArtist, false, false, 0)
	spotifyView.PackStart(spotifyAlbum, false, false, 0)
	spotifyView.PackStart(spotifyStatus, false, false, 0)

	spotifyList = gtk.NewListBox()
	spotifyList.SetName("spotify-list")
	spotifyList.SetSelectionMode(gtk.SelectionSingle)
	spotifyScroll = gtk.NewScrolledWindow(nil, nil)
	spotifyScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	spotifyScroll.SetNoShowAll(true)
	spotifyScroll.Add(spotifyList)

	for _, ctrl := range modules.PlayerControls(playerMode) {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}

	spotifyList.Connect("row-activated", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		idx := row.Index()
		ctrls := modules.PlayerControls(playerMode)
		if idx >= 0 && idx < len(ctrls) {
			runPlayerResultAction(ctrls[idx])
		}
	})

}

func createSpotifyControlRow(r modules.Result) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	hbox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	hbox.SetMarginStart(8)
	hbox.SetMarginEnd(8)
	hbox.SetMarginTop(6)
	hbox.SetMarginBottom(6)

	if icon := loadIcon(r.Icon, 20); icon != nil {
		hbox.PackStart(icon, false, false, 0)
	}

	label := gtk.NewLabel(r.Title)
	label.SetXAlign(0)
	hbox.PackStart(label, true, true, 0)

	if r.Desc != "" {
		desc := gtk.NewLabel(r.Desc)
		desc.SetXAlign(1)
		desc.SetName("edwin-desc")
		hbox.PackEnd(desc, false, false, 0)
	}

	row.Add(hbox)
	return row
}

func showSpotifyView() {
	showPlayerView(modules.PlayerSpotify)
}

func showPlayerView(kind modules.PlayerKind) {
	if inSpotifyMode && playerMode == kind {
		return
	}
	playerMode = kind
	inSpotifyMode = true
	resultsScroll.Hide()
	spotifyList.Show()
	spotifyScroll.Show()
	hidePreview()
	refreshPlayerControls()

	refreshSpotifyInfo()

	spotifyView.Show()
	spotifyView.ShowAll()
	spotifyScroll.ShowAll()

	if first := spotifyList.RowAtIndex(0); first != nil {
		spotifyList.SelectRow(first)
	}
}

func refreshPlayerControls() {
	if spotifyList == nil {
		return
	}
	for {
		row := spotifyList.RowAtIndex(0)
		if row == nil {
			break
		}
		spotifyList.Remove(row)
	}
	for _, ctrl := range modules.PlayerControls(playerMode) {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}
}

func hideSpotifyView() {
	inSpotifyMode = false
	if spotifyScroll != nil {
		spotifyScroll.Hide()
	}
	spotifyView.Hide()
}

func refreshSpotifyInfo() {
	mode := playerMode
	version := atomic.AddUint64(&playerInfoVersion, 1)
	go func() {
		info := modules.GetPlayerInfo(mode)
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&playerInfoVersion) != version {
				return
			}
			applyPlayerInfo(info)
		})
	}()
}

func applyPlayerInfo(info *modules.PlayerInfo) {
	if info == nil {
		showPlayerDisconnected()
		return
	}

	spotifyTitle.SetText(info.Title)
	spotifyArtist.SetText(info.Artist)
	spotifyAlbum.SetText(info.Album)
	spotifyStatus.SetText(playerStatusIcon(info.Status) + " " + string(info.Status))

	if info.ArtCachePath != "" {
		if pb, err := gdkpixbuf.NewPixbufFromFileAtSize(info.ArtCachePath, playerArtBigSize, playerArtBigSize); err == nil {
			spotifyArtBig.SetFromPixbuf(pb)
		}
	}
}

func runPlayerResultAction(result modules.Result) {
	go func() {
		executeActionSpec(result.ActionSpec)
		time.Sleep(playerRefreshDelay)
		glib.IdleAdd(func() { refreshSpotifyInfo() })
	}()
}

func showPlayerDisconnected() {
	if playerMode == modules.PlayerYouTube {
		spotifyTitle.SetText("No YouTube player detected")
		spotifyStatus.SetText("Open YouTube in browser")
	} else {
		spotifyTitle.SetText("No Spotify player detected")
		spotifyStatus.SetText("Start Spotify")
	}
	spotifyArtist.SetText("")
	spotifyAlbum.SetText("")
	spotifyArtBig.Clear()
}

func playerStatusIcon(status modules.PlaybackStatus) string {
	switch status {
	case modules.StatusPaused:
		return "⏸"
	case modules.StatusStopped:
		return "⏹"
	default:
		return "▶"
	}
}
