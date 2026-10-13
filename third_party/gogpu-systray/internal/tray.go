package internal

import "sync/atomic"

type TrayID uint32

var nextTrayID atomic.Uint32

func NewTrayID() TrayID {
	return TrayID(nextTrayID.Add(1))
}

type Tray struct {
	ID           TrayID
	Platform     PlatformTray
	Callbacks    Callbacks
	Icon         []byte
	DarkModeIcon []byte
	TemplateIcon []byte
	Tooltip      string
	AppName      string
	Menu         *Menu
	Visible      bool
}

func NewTray(platform PlatformTray) *Tray {
	return &Tray{
		ID:       NewTrayID(),
		Platform: platform,
	}
}

func (t *Tray) SetIcon(png []byte) error {
	t.Icon = png
	return t.Platform.SetIcon(png)
}

func (t *Tray) SetDarkModeIcon(png []byte) error {
	t.DarkModeIcon = png

	if setter, ok := t.Platform.(interface{ SetDarkModeIcon([]byte) error }); ok {
		return setter.SetDarkModeIcon(png)
	}

	return nil
}

func (t *Tray) SetTemplateIcon(png []byte) {
	t.TemplateIcon = png

	if setter, ok := t.Platform.(interface{ SetTemplateIcon([]byte) error }); ok {
		if err := setter.SetTemplateIcon(png); err != nil {

			_ = err
		}
	}
}

func (t *Tray) SetAppName(name string) {
	t.AppName = name
	if setter, ok := t.Platform.(interface{ SetAppName(string) }); ok {
		setter.SetAppName(name)
	}
}

func (t *Tray) SetTooltip(text string) error {
	t.Tooltip = text
	return t.Platform.SetTooltip(text)
}

func (t *Tray) SetMenu(menu *Menu) error {
	t.Menu = menu
	switch updater := t.Platform.(type) {
	case menuItemSnapshotUpdater:
		setMenuSnapshotUpdater(menu, updater)
	case MenuItemUpdater:
		SetMenuUpdater(menu, updater)
	}
	return t.Platform.SetMenu(menu)
}

func (t *Tray) Show() error {
	t.Visible = true
	return t.Platform.Show()
}

func (t *Tray) Hide() error {
	t.Visible = false
	return t.Platform.Hide()
}

func (t *Tray) ShowNotification(title, message string) error {
	return t.Platform.ShowNotification(title, message)
}

func (t *Tray) Bounds() (x, y, w, h int) {
	return t.Platform.Bounds()
}

func (t *Tray) Run() error {
	return t.Platform.Run()
}

func (t *Tray) Remove() {
	t.Visible = false
	t.Platform.Destroy()
}
