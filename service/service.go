package service

import "os"

// List all mist related service files in systemd directory.
func ListAllServices() (filenames []string, err error) {
	const dir = "/etc/systemd/system"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return filenames, err
	}

	for _, entry := range entries {
		if isMistFormat(entry.Name()) {
			filenames = append(filenames, entry.Name())
		}
	}

	return filenames, err
}
