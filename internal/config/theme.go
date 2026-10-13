package config

const (
	RadiusSmall  = 6
	RadiusMedium = 8

	SpacingTiny   = 4
	SpacingSmall  = 6
	SpacingMedium = 8
	SpacingLarge  = 12

	FontSizeSmall  = 11
	FontSizeBody   = 12
	FontSizeMedium = 13
	FontSizeTitle  = 14
	FontSizeLarge  = 16
	FontSizeIcon   = 22

	SpotifyControlSize = 36
)

type ColorPreset struct {
	BackgroundColor  string
	BackgroundAlpha  float64
	TextColor        string
	SelectionColor   string
	FocusBorderColor string
	EntryBackground  string
	DescColor        string
}

var ColorPresets = map[string]ColorPreset{
	"dark": {
		BackgroundColor:  "30, 30, 40",
		BackgroundAlpha:  0.95,
		TextColor:        "white",
		SelectionColor:   "100, 150, 255",
		FocusBorderColor: "100, 150, 255",
		EntryBackground:  "rgba(255, 255, 255, 0.1)",
		DescColor:        "rgba(255, 255, 255, 0.6)",
	},
	"midnight": {
		BackgroundColor:  "8, 8, 12",
		BackgroundAlpha:  0.97,
		TextColor:        "white",
		SelectionColor:   "40, 55, 95",
		FocusBorderColor: "40, 55, 95",
		EntryBackground:  "rgba(255, 255, 255, 0.07)",
		DescColor:        "rgba(255, 255, 255, 0.5)",
	},
	"light": {
		BackgroundColor:  "245, 245, 248",
		BackgroundAlpha:  0.97,
		TextColor:        "#1c1c1e",
		SelectionColor:   "90, 140, 255",
		FocusBorderColor: "90, 140, 255",
		EntryBackground:  "rgba(0, 0, 0, 0.06)",
		DescColor:        "rgba(0, 0, 0, 0.55)",
	},
	"classic": {
		BackgroundColor:  "238, 232, 213",
		BackgroundAlpha:  0.97,
		TextColor:        "#073642",
		SelectionColor:   "38, 139, 210",
		FocusBorderColor: "38, 139, 210",
		EntryBackground:  "rgba(0, 0, 0, 0.05)",
		DescColor:        "rgba(7, 54, 66, 0.6)",
	},
	"edwin": {
		BackgroundColor:  "24, 26, 34",
		BackgroundAlpha:  0.97,
		TextColor:        "#fdf2d6",
		SelectionColor:   "208, 193, 180",
		FocusBorderColor: "245, 215, 140",
		EntryBackground:  "rgba(208, 193, 180, 0.2)",
		DescColor:        "rgba(253, 242, 214, 0.62)",
	},
}

const StaticThemeCSS = `
	@define-color edwin-white rgba(255, 255, 255, 1);
	@define-color edwin-white-08 rgba(255, 255, 255, 0.08);
	@define-color edwin-white-10 rgba(255, 255, 255, 0.1);
	@define-color edwin-white-14 rgba(255, 255, 255, 0.14);
	@define-color edwin-white-20 rgba(255, 255, 255, 0.2);
	@define-color edwin-white-50 rgba(255, 255, 255, 0.5);
	@define-color edwin-white-60 rgba(255, 255, 255, 0.6);
	@define-color edwin-white-70 rgba(255, 255, 255, 0.7);
	@define-color edwin-white-80 rgba(255, 255, 255, 0.8);
	@define-color edwin-black-20 rgba(0, 0, 0, 0.2);
	@define-color edwin-black-30 rgba(0, 0, 0, 0.3);
	@define-color edwin-black-92 rgba(0, 0, 0, 0.92);
	@define-color edwin-spotify-green #1DB954;
`
