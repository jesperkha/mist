package service

import "github.com/godbus/dbus"

type Unit struct {
	ID          uint       `json:"id"`       // Same as in database
	Filename    string     `json:"filename"` // Full file name
	Name        string     `json:"name"`     // Just the service name
	Description string     `json:"description"`
	Status      UnitStatus `json:"status"`
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
