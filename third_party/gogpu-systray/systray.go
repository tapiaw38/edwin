package systray

import (
	"github.com/gogpu/systray/internal"
)

type TrayID uint32

type SystemTray struct {
	impl *internal.Tray
}

func New() *SystemTray {
	t := &SystemTray{}
	t.impl = &internal.Tray{
		ID: internal.NewTrayID(),
	}

	platform := internal.NewPlatformTray(&t.impl.Callbacks)
	t.impl.Platform = platform
	if err := platform.Create(); err != nil {
		return t
	}
	return t
}

func (t *SystemTray) ID() TrayID {
	return TrayID(t.impl.ID)
}

func (t *SystemTray) SetIcon(png []byte) *SystemTray {
	_ = t.impl.SetIcon(png)
	return t
}

func (t *SystemTray) SetDarkModeIcon(png []byte) *SystemTray {
	_ = t.impl.SetDarkModeIcon(png)
	return t
}

func (t *SystemTray) SetTemplateIcon(png []byte) *SystemTray {
	t.impl.SetTemplateIcon(png)
	return t
}

func (t *SystemTray) SetAppName(name string) *SystemTray {
	t.impl.SetAppName(name)
	return t
}

func (t *SystemTray) SetTooltip(text string) *SystemTray {
	_ = t.impl.SetTooltip(text)
	return t
}

func (t *SystemTray) SetMenu(menu *Menu) *SystemTray {
	var internalMenu *internal.Menu
	if menu != nil {
		internalMenu = menu.impl
	}
	_ = t.impl.SetMenu(internalMenu)
	return t
}

func (t *SystemTray) OnClick(fn func()) *SystemTray {
	t.impl.Callbacks.OnClick = fn
	return t
}

func (t *SystemTray) OnDoubleClick(fn func()) *SystemTray {
	t.impl.Callbacks.OnDoubleClick = fn
	return t
}

func (t *SystemTray) OnRightClick(fn func()) *SystemTray {
	t.impl.Callbacks.OnRightClick = fn
	return t
}

func (t *SystemTray) ShowNotification(title, message string) *SystemTray {
	_ = t.impl.ShowNotification(title, message)
	return t
}

func (t *SystemTray) Show() *SystemTray {
	_ = t.impl.Show()
	return t
}

func (t *SystemTray) Hide() *SystemTray {
	_ = t.impl.Hide()
	return t
}

func (t *SystemTray) Run() error {
	return t.impl.Run()
}

func (t *SystemTray) Remove() {
	t.impl.Remove()
}

func (t *SystemTray) Bounds() (x, y, w, h int) {
	return t.impl.Bounds()
}
