//go:build linux

package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"log/slog"
	"os"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	sniInterface     = "org.kde.StatusNotifierItem"
	sniPath          = "/StatusNotifierItem"
	menuInterface    = "com.canonical.dbusmenu"
	menuPath         = "/MenuBar"
	watcherInterface = "org.kde.StatusNotifierWatcher"
	watcherPath      = "/StatusNotifierWatcher"
	notifInterface   = "org.freedesktop.Notifications"
	notifPath        = "/org/freedesktop/Notifications"

	sniStatusPassive = "Passive"
	sniStatusActive  = "Active"

	dbusDirectionIn  = "in"
	dbusDirectionOut = "out"
)

type dbusPixmap struct {
	W    int32
	H    int32
	Data []byte
}

type dbusTooltip struct {
	IconName string
	Pixmaps  []dbusPixmap
	Title    string
	Text     string
}

type linuxTray struct {
	conn      *dbus.Conn
	props     *prop.Properties
	menuProps *prop.Properties
	busName   string
	trayID    TrayID
	callbacks *Callbacks

	menu      *Menu
	menuItems map[int32]*MenuItem
	itemIDs   map[uint32]int32
	menuRev   uint32

	iconPixmap []dbusPixmap
	tooltip    string
	appName    string
	status     string

	mu   sync.RWMutex
	quit chan struct{}
}

func NewPlatformTray(callbacks *Callbacks) PlatformTray {
	return &linuxTray{
		callbacks: callbacks,
		menuItems: make(map[int32]*MenuItem),
		itemIDs:   make(map[uint32]int32),
		status:    sniStatusPassive,
		quit:      make(chan struct{}),
	}
}

func (t *linuxTray) Create() error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("connect to session bus: %w", err)
	}
	t.conn = conn

	t.trayID = NewTrayID()
	t.busName = fmt.Sprintf("org.kde.StatusNotifierItem-%d-%d", os.Getpid(), t.trayID)
	reply, err := conn.RequestName(t.busName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return fmt.Errorf("request bus name %s: %w", t.busName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("bus name %s already taken (reply=%d)", t.busName, reply)
	}

	sniSvc := &sniService{tray: t}
	if err := conn.Export(sniSvc, sniPath, sniInterface); err != nil {
		return fmt.Errorf("export SNI service: %w", err)
	}

	sniProps, err := prop.Export(conn, sniPath, prop.Map{
		sniInterface: {
			"Category":      {Value: "ApplicationStatus", Writable: false, Emit: prop.EmitConst, Callback: nil},
			"Id":            {Value: t.busName, Writable: false, Emit: prop.EmitConst, Callback: nil},
			"Title":         {Value: "", Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"Status":        {Value: sniStatusPassive, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"IconName":      {Value: "", Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"IconPixmap":    {Value: []dbusPixmap{}, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"ToolTip":       {Value: dbusTooltip{}, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"Menu":          {Value: dbus.ObjectPath(menuPath), Writable: false, Emit: prop.EmitConst, Callback: nil},
			"ItemIsMenu":    {Value: false, Writable: false, Emit: prop.EmitConst, Callback: nil},
			"WindowId":      {Value: int32(0), Writable: false, Emit: prop.EmitConst, Callback: nil},
			"IconThemePath": {Value: "", Writable: false, Emit: prop.EmitConst, Callback: nil},
		},
	})
	if err != nil {
		return fmt.Errorf("export SNI properties: %w", err)
	}
	t.props = sniProps

	sniIntro := introspect.NewIntrospectable(&introspect.Node{
		Name: sniPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name: sniInterface,
				Methods: []introspect.Method{
					{Name: "Activate", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: dbusDirectionIn},
						{Name: "y", Type: "i", Direction: dbusDirectionIn},
					}},
					{Name: "SecondaryActivate", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: dbusDirectionIn},
						{Name: "y", Type: "i", Direction: dbusDirectionIn},
					}},
					{Name: "ContextMenu", Args: []introspect.Arg{
						{Name: "x", Type: "i", Direction: dbusDirectionIn},
						{Name: "y", Type: "i", Direction: dbusDirectionIn},
					}},
					{Name: "Scroll", Args: []introspect.Arg{
						{Name: "delta", Type: "i", Direction: dbusDirectionIn},
						{Name: "orientation", Type: "s", Direction: dbusDirectionIn},
					}},
				},
				Signals: []introspect.Signal{
					{Name: "NewTitle"},
					{Name: "NewIcon"},
					{Name: "NewToolTip"},
					{Name: "NewStatus", Args: []introspect.Arg{
						{Name: "status", Type: "s"},
					}},
				},
				Properties: sniProps.Introspection(sniInterface),
			},
		},
	})
	if err := conn.Export(sniIntro, sniPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return fmt.Errorf("export SNI introspection: %w", err)
	}

	menuSvc := &dbusMenuService{tray: t}
	if err := conn.Export(menuSvc, menuPath, menuInterface); err != nil {
		return fmt.Errorf("export dbusmenu service: %w", err)
	}

	menuPropMap, err := prop.Export(conn, menuPath, prop.Map{
		menuInterface: {
			"Version":       {Value: uint32(3), Writable: false, Emit: prop.EmitConst, Callback: nil},
			"TextDirection": {Value: "ltr", Writable: false, Emit: prop.EmitConst, Callback: nil},
			"Status":        {Value: "normal", Writable: false, Emit: prop.EmitConst, Callback: nil},
			"IconThemePath": {Value: []string{}, Writable: false, Emit: prop.EmitConst, Callback: nil},
		},
	})
	if err != nil {
		return fmt.Errorf("export dbusmenu properties: %w", err)
	}
	t.menuProps = menuPropMap

	menuIntro := introspect.NewIntrospectable(&introspect.Node{
		Name: menuPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name: menuInterface,
				Methods: []introspect.Method{
					{Name: "GetLayout", Args: []introspect.Arg{
						{Name: "parentId", Type: "i", Direction: dbusDirectionIn},
						{Name: "recursionDepth", Type: "i", Direction: dbusDirectionIn},
						{Name: "propertyNames", Type: "as", Direction: dbusDirectionIn},
						{Name: "revision", Type: "u", Direction: dbusDirectionOut},
						{Name: "layout", Type: "(ia{sv}av)", Direction: dbusDirectionOut},
					}},
					{Name: "GetGroupProperties", Args: []introspect.Arg{
						{Name: "ids", Type: "ai", Direction: dbusDirectionIn},
						{Name: "propertyNames", Type: "as", Direction: dbusDirectionIn},
						{Name: "properties", Type: "a(ia{sv})", Direction: dbusDirectionOut},
					}},
					{Name: "Event", Args: []introspect.Arg{
						{Name: "id", Type: "i", Direction: dbusDirectionIn},
						{Name: "eventId", Type: "s", Direction: dbusDirectionIn},
						{Name: "data", Type: "v", Direction: dbusDirectionIn},
						{Name: "timestamp", Type: "u", Direction: dbusDirectionIn},
					}},
					{Name: "AboutToShow", Args: []introspect.Arg{
						{Name: "id", Type: "i", Direction: dbusDirectionIn},
						{Name: "needUpdate", Type: "b", Direction: dbusDirectionOut},
					}},
					{Name: "EventGroup", Args: []introspect.Arg{
						{Name: "events", Type: "a(isvu)", Direction: dbusDirectionIn},
						{Name: "idErrors", Type: "ai", Direction: dbusDirectionOut},
					}},
					{Name: "AboutToShowGroup", Args: []introspect.Arg{
						{Name: "ids", Type: "ai", Direction: dbusDirectionIn},
						{Name: "updatesNeeded", Type: "ai", Direction: dbusDirectionOut},
						{Name: "idErrors", Type: "ai", Direction: dbusDirectionOut},
					}},
				},
				Signals: []introspect.Signal{
					{Name: "LayoutUpdated", Args: []introspect.Arg{
						{Name: "revision", Type: "u"},
						{Name: "parent", Type: "i"},
					}},
					{Name: "ItemsPropertiesUpdated", Args: []introspect.Arg{
						{Name: "updatedProps", Type: "a(ia{sv})"},
						{Name: "removedProps", Type: "a(ias)"},
					}},
				},
				Properties: menuPropMap.Introspection(menuInterface),
			},
		},
	})
	if err := conn.Export(menuIntro, menuPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return fmt.Errorf("export dbusmenu introspection: %w", err)
	}

	if err := t.registerWithWatcher(); err != nil {

		slog.Warn("systray: initial watcher registration failed (watcher may not be running)",
			"err", err)
	}

	go t.watchWatcherRestart()

	return nil
}

func (t *linuxTray) registerWithWatcher() error {
	obj := t.conn.Object(watcherInterface, watcherPath)
	call := obj.Call(watcherInterface+".RegisterStatusNotifierItem", 0, t.busName)
	if call.Err != nil {
		return fmt.Errorf("register status notifier item: %w", call.Err)
	}
	return nil
}

func (t *linuxTray) watchWatcherRestart() {
	if err := t.conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
		dbus.WithMatchArg(0, watcherInterface),
	); err != nil {
		slog.Warn("systray: failed to add NameOwnerChanged match", "err", err)
		return
	}

	ch := make(chan *dbus.Signal, 4)
	t.conn.Signal(ch)
	defer t.conn.RemoveSignal(ch)

	for {
		select {
		case sig, ok := <-ch:
			if !ok {
				return
			}
			if sig.Name != "org.freedesktop.DBus.NameOwnerChanged" {
				continue
			}

			if len(sig.Body) < 3 {
				continue
			}
			newOwner, ok := sig.Body[2].(string)
			if !ok || newOwner == "" {

				continue
			}

			if err := t.registerWithWatcher(); err != nil {
				slog.Warn("systray: re-registration with watcher failed", "err", err)
			}

		case <-t.quit:
			return
		}
	}
}

func (t *linuxTray) SetIcon(png []byte) error {
	if len(png) == 0 {
		return fmt.Errorf("empty icon data")
	}

	w, h, argb, err := pngToARGB(png)
	if err != nil {
		return fmt.Errorf("convert PNG to ARGB: %w", err)
	}

	t.mu.Lock()
	t.iconPixmap = []dbusPixmap{{W: int32(w), H: int32(h), Data: argb}}
	t.mu.Unlock()

	if t.props != nil {
		t.props.SetMust(sniInterface, "IconPixmap", t.iconPixmap)
		if err := t.emitSignal(sniInterface + ".NewIcon"); err != nil {
			slog.Warn("systray: emit NewIcon failed", "err", err)
		}
	}

	return nil
}

func (t *linuxTray) SetTooltip(text string) error {
	t.mu.Lock()
	t.tooltip = text
	t.mu.Unlock()

	if t.props != nil {
		t.props.SetMust(sniInterface, "Title", text)
		t.props.SetMust(sniInterface, "ToolTip", dbusTooltip{
			IconName: "",
			Pixmaps:  nil,
			Title:    text,
			Text:     "",
		})
		if err := t.emitSignal(sniInterface + ".NewTitle"); err != nil {
			slog.Warn("systray: emit NewTitle failed", "err", err)
		}
	}

	return nil
}

func (t *linuxTray) SetMenu(menu *Menu) error {
	t.mu.Lock()
	t.menu = menu
	t.menuItems = make(map[int32]*MenuItem)
	t.itemIDs = make(map[uint32]int32)
	if menu != nil {
		t.buildMenuItemMap(menu.Items, 1)
	}
	t.menuRev++
	rev := t.menuRev
	t.mu.Unlock()

	if t.conn != nil {
		if err := t.conn.Emit(menuPath, menuInterface+".LayoutUpdated", rev, int32(0)); err != nil {
			slog.Warn("systray: emit LayoutUpdated failed", "err", err)
		}
	}

	return nil
}

func (t *linuxTray) buildMenuItemMap(items []*MenuItem, nextID int32) int32 {
	for _, item := range items {
		id := nextID
		t.menuItems[id] = item
		t.itemIDs[item.ID()] = id
		nextID++
		if item.Type == MenuItemSubmenu && item.Submenu != nil {
			nextID = t.buildMenuItemMap(item.Submenu.Items, nextID)
		}
	}
	return nextID
}

func (t *linuxTray) UpdateItem(item *MenuItem) error {
	return t.updateItem(item.snapshot())
}

func (t *linuxTray) updateItem(item menuItemSnapshot) error {
	if t.conn == nil {
		return nil
	}

	t.mu.RLock()
	dbusID, ok := t.itemIDs[item.id]
	t.mu.RUnlock()
	if !ok {
		return nil
	}

	props := make(map[string]dbus.Variant)
	props["label"] = dbus.MakeVariant(item.label)
	if item.disabled {
		props["enabled"] = dbus.MakeVariant(false)
	} else {
		props["enabled"] = dbus.MakeVariant(true)
	}
	if item.itemType == MenuItemCheckbox {
		if item.checked {
			props["toggle-state"] = dbus.MakeVariant(int32(1))
		} else {
			props["toggle-state"] = dbus.MakeVariant(int32(0))
		}
	}
	if len(item.icon) > 0 {
		props["icon-data"] = dbus.MakeVariant(item.icon)
	}

	updated := []menuItemProps{{ID: dbusID, Props: props}}
	removed := []menuItemRemovedProps{}

	return t.conn.Emit(menuPath, menuInterface+".ItemsPropertiesUpdated", updated, removed)
}

func (t *linuxTray) SetAppName(name string) {
	t.appName = name
}

func (t *linuxTray) ShowNotification(title, message string) error {
	if t.conn == nil {
		return fmt.Errorf("dbus connection not established")
	}

	obj := t.conn.Object(notifInterface, notifPath)
	call := obj.Call(notifInterface+".Notify", 0,
		t.appName,
		uint32(0),
		"",
		title,
		message,
		[]string{},
		map[string]dbus.Variant{},
		int32(-1),
	)
	if call.Err != nil {
		return fmt.Errorf("notify: %w", call.Err)
	}

	return nil
}

func (t *linuxTray) Show() error {
	t.mu.Lock()
	t.status = sniStatusActive
	t.mu.Unlock()

	if t.props != nil {
		t.props.SetMust(sniInterface, "Status", sniStatusActive)
		if err := t.emitSignal(sniInterface+".NewStatus", sniStatusActive); err != nil {
			slog.Warn("systray: emit NewStatus failed", "err", err)
		}
	}

	return nil
}

func (t *linuxTray) Hide() error {
	t.mu.Lock()
	t.status = sniStatusPassive
	t.mu.Unlock()

	if t.props != nil {
		t.props.SetMust(sniInterface, "Status", sniStatusPassive)
		if err := t.emitSignal(sniInterface+".NewStatus", sniStatusPassive); err != nil {
			slog.Warn("systray: emit NewStatus failed", "err", err)
		}
	}

	return nil
}

func (t *linuxTray) Bounds() (int, int, int, int) {
	return 0, 0, 0, 0
}

func (t *linuxTray) Run() error {
	<-t.quit
	return nil
}

func (t *linuxTray) Destroy() {

	if t.conn != nil && t.props != nil {
		t.props.SetMust(sniInterface, "Status", sniStatusPassive)
		if err := t.emitSignal(sniInterface+".NewStatus", sniStatusPassive); err != nil {
			slog.Warn("systray: emit NewStatus (Passive) on destroy failed", "err", err)
		}
	}

	if t.conn != nil {
		if _, err := t.conn.ReleaseName(t.busName); err != nil {
			slog.Warn("systray: ReleaseName failed", "err", err)
		}
		if err := t.conn.Close(); err != nil {
			slog.Warn("systray: Close connection failed", "err", err)
		}
	}

	select {
	case <-t.quit:

	default:
		close(t.quit)
	}
}

func (t *linuxTray) emitSignal(name string, values ...interface{}) error {
	return t.conn.Emit(sniPath, name, values...)
}

type sniService struct {
	tray *linuxTray
}

func (s *sniService) Activate(x, y int32) *dbus.Error {
	if s.tray.callbacks != nil && s.tray.callbacks.OnClick != nil {
		s.tray.callbacks.OnClick()
	}
	return nil
}

func (s *sniService) SecondaryActivate(x, y int32) *dbus.Error {
	if s.tray.callbacks != nil && s.tray.callbacks.OnDoubleClick != nil {
		s.tray.callbacks.OnDoubleClick()
	}
	return nil
}

func (s *sniService) ContextMenu(x, y int32) *dbus.Error {
	if s.tray.callbacks != nil && s.tray.callbacks.OnRightClick != nil {
		s.tray.callbacks.OnRightClick()
	}
	return nil
}

func (s *sniService) Scroll(delta int32, orientation string) *dbus.Error {

	return nil
}

type menuLayout struct {
	V0 int32
	V1 map[string]dbus.Variant
	V2 []dbus.Variant
}

type menuItemProps struct {
	ID    int32
	Props map[string]dbus.Variant
}

type menuItemRemovedProps struct {
	ID    int32
	Props []string
}

type menuEvent struct {
	ID        int32
	EventID   string
	Data      dbus.Variant
	Timestamp uint32
}

type dbusMenuService struct {
	tray *linuxTray
}

func (m *dbusMenuService) GetLayout(parentID int32, recursionDepth int32, propertyNames []string) (uint32, menuLayout, *dbus.Error) {
	m.tray.mu.RLock()
	defer m.tray.mu.RUnlock()

	rev := m.tray.menuRev
	layout := m.buildLayout(parentID, recursionDepth, 0)

	return rev, layout, nil
}

func (m *dbusMenuService) buildLayout(id int32, maxDepth int32, currentDepth int32) menuLayout {
	if id == 0 {

		rootProps := map[string]dbus.Variant{
			"children-display": dbus.MakeVariant("submenu"),
		}
		var children []dbus.Variant
		if maxDepth != 0 && m.tray.menu != nil {
			children = m.buildChildren(m.tray.menu.Items, 1, maxDepth, currentDepth+1)
		}
		return menuLayout{V0: 0, V1: rootProps, V2: children}
	}

	item, ok := m.tray.menuItems[id]
	if !ok {
		return menuLayout{V0: id, V1: map[string]dbus.Variant{}, V2: nil}
	}

	props := m.itemProperties(item.snapshot())
	var children []dbus.Variant
	if item.Type == MenuItemSubmenu && item.Submenu != nil && maxDepth != 0 {

		childStartID := m.findChildStartID(id)
		if childStartID > 0 {
			children = m.buildChildren(item.Submenu.Items, childStartID, maxDepth, currentDepth+1)
		}
	}

	return menuLayout{V0: id, V1: props, V2: children}
}

func (m *dbusMenuService) buildChildren(items []*MenuItem, startID int32, maxDepth int32, currentDepth int32) []dbus.Variant {
	children := make([]dbus.Variant, 0, len(items))
	nextID := startID

	for _, item := range items {
		id := nextID
		nextID++

		props := m.itemProperties(item.snapshot())
		var subChildren []dbus.Variant

		if item.Type == MenuItemSubmenu && item.Submenu != nil {
			if maxDepth < 0 || currentDepth < maxDepth {
				subChildren = m.buildChildren(item.Submenu.Items, nextID, maxDepth, currentDepth+1)
			}

			nextID = m.advancePastSubmenu(item.Submenu, nextID)
		}

		child := menuLayout{V0: id, V1: props, V2: subChildren}
		children = append(children, dbus.MakeVariant(child))
	}

	return children
}

func (m *dbusMenuService) advancePastSubmenu(menu *Menu, startID int32) int32 {
	nextID := startID
	for _, item := range menu.Items {
		nextID++
		if item.Type == MenuItemSubmenu && item.Submenu != nil {
			nextID = m.advancePastSubmenu(item.Submenu, nextID)
		}
	}
	return nextID
}

func (m *dbusMenuService) findChildStartID(parentID int32) int32 {

	return parentID + 1
}

func (m *dbusMenuService) itemProperties(item menuItemSnapshot) map[string]dbus.Variant {
	props := make(map[string]dbus.Variant)

	switch item.itemType {
	case MenuItemSeparator:
		props["type"] = dbus.MakeVariant("separator")

	case MenuItemCheckbox:
		props["label"] = dbus.MakeVariant(item.label)
		props["toggle-type"] = dbus.MakeVariant("checkmark")
		if item.checked {
			props["toggle-state"] = dbus.MakeVariant(int32(1))
		} else {
			props["toggle-state"] = dbus.MakeVariant(int32(0))
		}
		if item.disabled {
			props["enabled"] = dbus.MakeVariant(false)
		}

	case MenuItemSubmenu:
		props["label"] = dbus.MakeVariant(item.label)
		props["children-display"] = dbus.MakeVariant("submenu")
		if item.disabled {
			props["enabled"] = dbus.MakeVariant(false)
		}

	default:
		props["label"] = dbus.MakeVariant(item.label)
		if item.disabled {
			props["enabled"] = dbus.MakeVariant(false)
		}
		if len(item.icon) > 0 {
			props["icon-data"] = dbus.MakeVariant(item.icon)
		}
	}

	return props
}

func (m *dbusMenuService) GetGroupProperties(ids []int32, propertyNames []string) ([]menuItemProps, *dbus.Error) {
	m.tray.mu.RLock()
	defer m.tray.mu.RUnlock()

	result := make([]menuItemProps, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			result = append(result, menuItemProps{
				ID: 0,
				Props: map[string]dbus.Variant{
					"children-display": dbus.MakeVariant("submenu"),
				},
			})
			continue
		}
		item, ok := m.tray.menuItems[id]
		if !ok {
			continue
		}
		result = append(result, menuItemProps{
			ID:    id,
			Props: m.itemProperties(item.snapshot()),
		})
	}

	return result, nil
}

func (m *dbusMenuService) Event(id int32, eventID string, data dbus.Variant, timestamp uint32) *dbus.Error {
	if eventID != "clicked" {
		return nil
	}

	m.tray.mu.RLock()
	item, ok := m.tray.menuItems[id]
	m.tray.mu.RUnlock()

	if !ok {
		return nil
	}

	if item.OnClick != nil {
		item.OnClick()
	}

	return nil
}

func (m *dbusMenuService) AboutToShow(id int32) (bool, *dbus.Error) {
	return false, nil
}

func (m *dbusMenuService) EventGroup(events []menuEvent) ([]int32, *dbus.Error) {
	var idErrors []int32
	for _, ev := range events {
		if err := m.Event(ev.ID, ev.EventID, ev.Data, ev.Timestamp); err != nil {
			idErrors = append(idErrors, ev.ID)
		}
	}
	if idErrors == nil {
		idErrors = []int32{}
	}
	return idErrors, nil
}

func (m *dbusMenuService) AboutToShowGroup(ids []int32) ([]int32, []int32, *dbus.Error) {

	return []int32{}, []int32{}, nil
}

func pngToARGB(data []byte) (w, h int, argb []byte, err error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, 0, nil, fmt.Errorf("decode PNG: %w", err)
	}

	bounds := img.Bounds()
	w = bounds.Dx()
	h = bounds.Dy()
	argb = make([]byte, w*h*4)

	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			binary.BigEndian.PutUint32(argb[idx:], (a>>8)<<24|(r>>8)<<16|(g>>8)<<8|(b>>8))
			idx += 4
		}
	}

	return w, h, argb, nil
}
