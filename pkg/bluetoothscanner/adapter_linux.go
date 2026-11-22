package bluetoothscanner

import (
	"strings"

	"tinygo.org/x/bluetooth"
)

func selectAdapter(id string) (*bluetooth.Adapter, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return bluetooth.DefaultAdapter, nil
	}
	return bluetooth.NewAdapter(id), nil
}
