package systray

import (
	"github.com/gogpu/systray/internal"
)

type MenuItemType = internal.MenuItemType

const (
	MenuItemNormal    = internal.MenuItemNormal
	MenuItemCheckbox  = internal.MenuItemCheckbox
	MenuItemSeparator = internal.MenuItemSeparator
	MenuItemSubmenu   = internal.MenuItemSubmenu
)

type MenuItem struct {
	impl *internal.MenuItem
}

func (item *MenuItem) IsChecked() bool {
	return item.impl.IsChecked()
}

func (item *MenuItem) IsDisabled() bool {
	return item.impl.IsDisabled()
}

func (item *MenuItem) SetLabel(label string) {
	item.impl.SetLabel(label)
}

func (item *MenuItem) SetChecked(checked bool) {
	item.impl.SetChecked(checked)
}

func (item *MenuItem) SetDisabled(disabled bool) {
	item.impl.SetDisabled(disabled)
}

func (item *MenuItem) SetIcon(png []byte) {
	item.impl.SetIcon(png)
}

type Menu struct {
	impl *internal.Menu
}

func NewMenu() *Menu {
	return &Menu{impl: internal.NewMenu()}
}

func (m *Menu) Add(label string, onClick func()) *MenuItem {
	item := m.impl.Add(label, onClick)
	return &MenuItem{impl: item}
}

func (m *Menu) AddCheckbox(label string, checked bool, onClick func()) *MenuItem {
	item := m.impl.AddCheckbox(label, checked, onClick)
	return &MenuItem{impl: item}
}

func (m *Menu) AddSeparator() *Menu {
	m.impl.AddSeparator()
	return m
}

func (m *Menu) AddSubmenu(label string, submenu *Menu) *MenuItem {
	item := m.impl.AddSubmenu(label, submenu.impl)
	return &MenuItem{impl: item}
}

func (m *Menu) AddWithIcon(label string, icon []byte, onClick func()) *MenuItem {
	item := m.impl.AddWithIcon(label, icon, onClick)
	return &MenuItem{impl: item}
}
