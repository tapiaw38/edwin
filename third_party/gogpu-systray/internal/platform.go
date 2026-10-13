package internal

type PlatformTray interface {
	Create() error
	SetIcon(png []byte) error
	SetTooltip(text string) error
	SetMenu(menu *Menu) error
	ShowNotification(title, message string) error
	Show() error
	Hide() error
	Bounds() (x, y, w, h int)
	Run() error
	Destroy()
}

type MenuItemUpdater interface {
	UpdateItem(item *MenuItem) error
}

type menuItemSnapshotUpdater interface {
	updateItem(item menuItemSnapshot) error
}

type Callbacks struct {
	OnClick       func()
	OnDoubleClick func()
	OnRightClick  func()
}
