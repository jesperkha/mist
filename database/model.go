package database

type Service struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`        // Name of service. It must be the same as the service file.
	Description string `json:"description"` // Short description
	Port        string `json:"port"`        // Port service is running at, without colon.
	RequireAuth bool   `json:"requireAuth"` // If true, add auth middleware to proxy handler.
}
