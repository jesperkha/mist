package database

type Service struct {
	ID          uint
	Name        string // Name of service. It must be the same as the service file.
	Description string // Short description
	Port        string // Port service is running at
}
