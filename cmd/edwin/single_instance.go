package main

import "github.com/godbus/dbus/v5"

const (
	instanceBusName = "com.tapiaw38.Edwin"
	instancePath    = dbus.ObjectPath("/com/tapiaw38/Edwin")
	instanceIface   = "com.tapiaw38.Edwin"
)

type activationHandler struct {
	activate func()
}

func (h activationHandler) Activate() *dbus.Error {
	h.activate()
	return nil
}

func claimSingleInstance(activate func()) (cleanup func(), alreadyRunning bool, err error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, false, err
	}

	if err := conn.Export(activationHandler{activate: activate}, instancePath, instanceIface); err != nil {
		_ = conn.Close()
		return nil, false, err
	}

	reply, err := conn.RequestName(instanceBusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		_ = conn.Close()
		return nil, false, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner && reply != dbus.RequestNameReplyAlreadyOwner {
		callErr := conn.Object(instanceBusName, instancePath).Call(instanceIface+".Activate", 0).Err
		_ = conn.Close()
		return nil, true, callErr
	}

	return func() {
		_, _ = conn.ReleaseName(instanceBusName)
		_ = conn.Close()
	}, false, nil
}
