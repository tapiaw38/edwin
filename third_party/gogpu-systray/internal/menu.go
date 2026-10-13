package internal

import (
	"sync"
	"sync/atomic"
)

type MenuItemType int

const (
	MenuItemNormal MenuItemType = iota
	MenuItemCheckbox
	MenuItemSeparator
	MenuItemSubmenu
)

var nextMenuItemID atomic.Uint32

func newMenuItemID() uint32 {
	return nextMenuItemID.Add(1)
}

type MenuItem struct {
	Label    string
	Icon     []byte
	Type     MenuItemType
	Checked  bool
	Disabled bool
	Submenu  *Menu
	OnClick  func()

	id      uint32
	mu      sync.Mutex
	updater menuItemSnapshotUpdater
}

type menuItemSnapshot struct {
	id       uint32
	label    string
	icon     []byte
	itemType MenuItemType
	checked  bool
	disabled bool
}

func (item *MenuItem) ID() uint32 { return item.id }

func (item *MenuItem) snapshotLocked() menuItemSnapshot {
	return menuItemSnapshot{
		id:       item.id,
		label:    item.Label,
		icon:     append([]byte(nil), item.Icon...),
		itemType: item.Type,
		checked:  item.Checked,
		disabled: item.Disabled,
	}
}

func (item *MenuItem) updateLocked() (menuItemSnapshotUpdater, menuItemSnapshot) {
	if item.updater == nil {
		return nil, menuItemSnapshot{}
	}
	return item.updater, item.snapshotLocked()
}

func (item *MenuItem) snapshot() menuItemSnapshot {
	item.mu.Lock()
	snapshot := item.snapshotLocked()
	item.mu.Unlock()
	return snapshot
}

func (item *MenuItem) IsChecked() bool {
	item.mu.Lock()
	v := item.Checked
	item.mu.Unlock()
	return v
}

func (item *MenuItem) IsDisabled() bool {
	item.mu.Lock()
	v := item.Disabled
	item.mu.Unlock()
	return v
}

func (item *MenuItem) SetLabel(label string) {
	item.mu.Lock()
	item.Label = label
	u, snapshot := item.updateLocked()
	item.mu.Unlock()
	if u != nil {
		_ = u.updateItem(snapshot)
	}
}

func (item *MenuItem) SetChecked(checked bool) {
	item.mu.Lock()
	item.Checked = checked
	u, snapshot := item.updateLocked()
	item.mu.Unlock()
	if u != nil {
		_ = u.updateItem(snapshot)
	}
}

func (item *MenuItem) SetDisabled(disabled bool) {
	item.mu.Lock()
	item.Disabled = disabled
	u, snapshot := item.updateLocked()
	item.mu.Unlock()
	if u != nil {
		_ = u.updateItem(snapshot)
	}
}

func (item *MenuItem) SetIcon(png []byte) {
	item.mu.Lock()
	item.Icon = append([]byte(nil), png...)
	u, snapshot := item.updateLocked()
	item.mu.Unlock()
	if u != nil {
		_ = u.updateItem(snapshot)
	}
}

func (item *MenuItem) setUpdater(u menuItemSnapshotUpdater) {
	item.mu.Lock()
	item.updater = u
	item.mu.Unlock()
}

func (item *MenuItem) SetUpdater(u MenuItemUpdater) {
	if u == nil {
		item.setUpdater(nil)
		return
	}
	item.setUpdater(legacyMenuItemUpdater{updater: u, item: item})
}

type Menu struct {
	Items []*MenuItem
}

func NewMenu() *Menu {
	return &Menu{}
}

func (m *Menu) Add(label string, onClick func()) *MenuItem {
	item := &MenuItem{
		id:      newMenuItemID(),
		Label:   label,
		Type:    MenuItemNormal,
		OnClick: onClick,
	}
	m.Items = append(m.Items, item)
	return item
}

func (m *Menu) AddCheckbox(label string, checked bool, onClick func()) *MenuItem {
	item := &MenuItem{
		id:      newMenuItemID(),
		Label:   label,
		Type:    MenuItemCheckbox,
		Checked: checked,
		OnClick: onClick,
	}
	m.Items = append(m.Items, item)
	return item
}

func (m *Menu) AddSeparator() *MenuItem {
	item := &MenuItem{
		id:   newMenuItemID(),
		Type: MenuItemSeparator,
	}
	m.Items = append(m.Items, item)
	return item
}

func (m *Menu) AddSubmenu(label string, submenu *Menu) *MenuItem {
	item := &MenuItem{
		id:      newMenuItemID(),
		Label:   label,
		Type:    MenuItemSubmenu,
		Submenu: submenu,
	}
	m.Items = append(m.Items, item)
	return item
}

func (m *Menu) AddWithIcon(label string, icon []byte, onClick func()) *MenuItem {
	item := &MenuItem{
		id:      newMenuItemID(),
		Label:   label,
		Icon:    append([]byte(nil), icon...),
		Type:    MenuItemNormal,
		OnClick: onClick,
	}
	m.Items = append(m.Items, item)
	return item
}

func SetMenuUpdater(menu *Menu, u MenuItemUpdater) {
	if menu == nil {
		return
	}
	for _, item := range menu.Items {
		item.SetUpdater(u)
		if item.Type == MenuItemSubmenu && item.Submenu != nil {
			SetMenuUpdater(item.Submenu, u)
		}
	}
}

type legacyMenuItemUpdater struct {
	updater MenuItemUpdater
	item    *MenuItem
}

func (u legacyMenuItemUpdater) updateItem(menuItemSnapshot) error {
	return u.updater.UpdateItem(u.item)
}

func setMenuSnapshotUpdater(menu *Menu, u menuItemSnapshotUpdater) {
	if menu == nil {
		return
	}
	for _, item := range menu.Items {
		item.setUpdater(u)
		if item.Type == MenuItemSubmenu && item.Submenu != nil {
			setMenuSnapshotUpdater(item.Submenu, u)
		}
	}
}
