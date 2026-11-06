module github.com/tapiaw38/edwin

go 1.26.1

require (
	github.com/diamondburned/gotk4-layer-shell/pkg v0.0.0-20240109211357-6efa9f6dc438
	github.com/diamondburned/gotk4/pkg v0.4.1
	github.com/godbus/dbus/v5 v5.2.2
	github.com/gogpu/systray v0.3.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/gogpu/systray => ./third_party/gogpu-systray

require (
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.47.0 // indirect
)
