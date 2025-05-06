package service

import (
	"log"
	"strings"

	"github.com/godbus/dbus"
	"github.com/jesperkha/mist/config"
)

type Monitor struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

type Unit struct {
	Filename    string // Full file name
	Name        string // Just the service name
	Description string
	Status      UnitStatus
}

type UnitStatus int

const (
	Stopped UnitStatus = iota
	Running
	Crashed
)

func (s UnitStatus) String() string {
	return map[UnitStatus]string{
		Stopped: "stopped",
		Running: "running",
		Crashed: "crashed",
	}[s]
}

type dbusUnit struct {
	Name        string
	Description string
	LoadState   string
	ActiveState string
	SubState    string
	Followed    dbus.ObjectPath
	Path        dbus.ObjectPath
	JobId       uint32
	JobType     string
	JobPath     dbus.ObjectPath
}

func NewMonitor(config *config.Config) *Monitor {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("dbus conn: %v", err)
	}

	return &Monitor{
		conn: conn,
		obj:  conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1"),
	}
}

func (m *Monitor) Close() error {
	return m.conn.Close()
}

// Poll all daemons assosiated with mist (correct service file name format).
func (m *Monitor) Poll() (units []Unit, err error) {
	var allUnits []dbusUnit
	if err := m.obj.Call("org.freedesktop.systemd1.Manager.ListUnits", 0).Store(&allUnits); err != nil {
		return units, err
	}

	for _, u := range allUnits {
		if !isMistFormat(u.Name) {
			continue
		}

		unit := Unit{
			Name:        serviceName(u.Name),
			Filename:    u.Name,
			Description: u.Description,
			Status:      status(u),
		}

		units = append(units, unit)
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
	filename := PREFIX + name + SUFFIX
	var jobPath dbus.ObjectPath
	return m.obj.Call("org.freedesktop.systemd1.Manager."+cmd, 0, filename, "replace").Store(&jobPath)
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

// The mist format for service files is:
//		mistservice_<name>.service

const PREFIX = "mistservice_"
const SUFFIX = ".service"

func isMistFormat(filename string) bool {
	return strings.HasPrefix(filename, PREFIX) && strings.HasSuffix(filename, SUFFIX)
}

func serviceName(filename string) string {
	s, _ := strings.CutPrefix(filename, PREFIX)
	s, _ = strings.CutSuffix(s, SUFFIX)
	return s
}
