package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus"
	"github.com/jesperkha/mist/database"
)

type Monitor struct {
	conn *dbus.Conn
	db   *database.Database
	obj  dbus.BusObject
}

func NewMonitor(db *database.Database) *Monitor {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("dbus conn: %v", err)
	}

	return &Monitor{
		conn: conn,
		db:   db,
		obj:  conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1"),
	}
}

func (m *Monitor) CloseConn() error {
	return m.conn.Close()
}

// Poll returns a list of all 'live' services handled by systemd.
func (m *Monitor) Poll() (units []Unit, err error) {
	var allUnits []dbusUnit
	if err := m.obj.Call("org.freedesktop.systemd1.Manager.ListUnits", 0).Store(&allUnits); err != nil {
		return nil, err
	}

	unitMap := make(map[string]dbusUnit)
	for _, u := range allUnits {
		name := serviceName(u.Name)
		unitMap[name] = u
	}

	services, err := m.db.GetAllServices()
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		stat := Stopped
		if unit, ok := unitMap[s.Name]; ok {
			stat = status(unit)
		}

		units = append(units, Unit{
			Service: s,
			Status:  stat,
		})
	}

	return units, err
}

// StartService takes a plain servie name (Unit.Name) and queries systemd to start it.
// Requires root permissions.
func (m *Monitor) StartService(name string) error {
	return m.controlService(name, "StartUnit")
}

// StopService takes a plain servie name (Unit.Name) and queries systemd to stop it.
// Requires root permissions.
func (m *Monitor) StopService(name string) error {
	return m.controlService(name, "StopUnit")
}

// cmd is either StartUnit or StopUnit
func (m *Monitor) controlService(name string, cmd string) error {
	filename := fmt.Sprintf("%s.service", name)
	var jobPath dbus.ObjectPath
	return m.obj.Call("org.freedesktop.systemd1.Manager."+cmd, 0, filename, "replace").Store(&jobPath)
}

func serviceName(filename string) string {
	s, _ := strings.CutSuffix(filename, ".service")
	return s
}

func status(u dbusUnit) UnitStatus {
	active, load, sub := u.ActiveState, u.LoadState, u.SubState
	if active == "active" && (sub == "running" || sub == "exited") {
		return Running
	}
	if active == "failed" || sub == "failed" || load == "error" || load == "not-found" {
		return Crashed
	}
	return Stopped
}
