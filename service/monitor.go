package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus"
	"github.com/jesperkha/mist/config"
	"github.com/jesperkha/mist/database"
)

type Monitor struct {
	conn *dbus.Conn
	db   *database.Database
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

func NewMonitor(config *config.Config, db *database.Database) *Monitor {
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

func (m *Monitor) Close() error {
	return m.conn.Close()
}

func (m *Monitor) serviceNameMap() (map[string]struct{}, error) {
	services := make(map[string]struct{})

	all, err := m.db.GetAllServices()
	if err != nil {
		return nil, err
	}

	for _, s := range all {
		services[s.Name] = struct{}{}
	}

	return services, nil
}

// Poll returns a list of all services in the database handled by systemd.
func (m *Monitor) Poll() (units []Unit, err error) {
	var allUnits []dbusUnit
	if err := m.obj.Call("org.freedesktop.systemd1.Manager.ListUnits", 0).Store(&allUnits); err != nil {
		return units, err
	}

	smap, err := m.serviceNameMap()
	if err != nil {
		return units, err
	}

	for _, u := range allUnits {
		name := serviceName(u.Name)
		if _, ok := smap[name]; !ok {
			continue
		}

		unit := Unit{
			Name:        name,
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
