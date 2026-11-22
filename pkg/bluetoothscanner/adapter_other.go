//go:build !linux

package bluetoothscanner

import (
	"errors"
	"strings"

	"tinygo.org/x/bluetooth"
)

func selectAdapter(id string) (*bluetooth.Adapter, error) {
	if strings.TrimSpace(id) != "" {
		return nil, errors.New("Bluetooth adapter selection is only supported on Linux")
	}
	return bluetooth.DefaultAdapter, nil
}
