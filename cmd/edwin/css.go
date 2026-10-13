package main

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/edwin/internal/config"
)

func loadCSS() {
	css := gtk.NewCSSProvider()
	css.LoadFromData(config.GetCSS() + fmt.Sprintf(`
		#edwin-preview {
			background: @edwin-black-30;
			padding: %[2]dpx;
			border-radius: %[1]dpx;
		}
		#edwin-preview-label {
			color: @edwin-white-80;
			font-family: monospace;
			font-size: %[3]dpx;
		}
		#edwin-preview-image {
			background: @edwin-white-08;
			border: 1px solid @edwin-white-14;
			padding: %[2]dpx;
		}
		#edwin-icon-text {
			font-family: "Noto Color Emoji", "Twemoji", "Segoe UI Emoji", sans-serif;
			font-size: %[4]dpx;
		}
		#spotify-view {
			background: @edwin-black-20;
			border-radius: %[5]dpx;
			padding: %[6]dpx;
		}
		#spotify-header {
			background: @edwin-black-30;
			border-radius: %[5]dpx;
			padding: %[6]dpx;
		}
		#spotify-title {
			color: @edwin-white;
			font-size: %[7]dpx;
			font-weight: bold;
		}
		#spotify-artist {
			color: @edwin-white-70;
			font-size: %[8]dpx;
		}
		#spotify-album {
			color: @edwin-white-50;
			font-size: %[10]dpx;
		}
		#spotify-status {
			color: @edwin-spotify-green;
			font-size: %[3]dpx;
		}
		#spotify-control {
			background: @edwin-white-10;
			border-radius: 50%%;
			padding: %[2]dpx;
			min-width: %[9]dpx;
			min-height: %[9]dpx;
		}
		#spotify-control:hover {
			background: @edwin-white-20;
		}
		#spotify-list {
			background: transparent;
		}
		#spotify-list row {
			background: transparent;
			border-radius: %[1]dpx;
			padding: %[1]dpx;
		}
		#spotify-list row:selected {
			background: @edwin-selection;
		}
	`, config.RadiusSmall, config.SpacingMedium, config.FontSizeSmall,
		config.FontSizeIcon, config.RadiusMedium, config.SpacingLarge,
		config.FontSizeLarge, config.FontSizeMedium, config.SpotifyControlSize,
		config.FontSizeBody))
	screen := gdk.ScreenGetDefault()
	gtk.StyleContextAddProviderForScreen(screen, css, uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))
}
