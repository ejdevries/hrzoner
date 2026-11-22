package bluetoothscanner

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

type Options struct {
	AdapterID string
	Duration  time.Duration
}

type Device struct {
	Address          string    `json:"address"`
	AddressType      string    `json:"addressType"`
	RSSI             int16     `json:"rssi"`
	Name             string    `json:"name"`
	ServiceUUIDs     []string  `json:"serviceUuids"`
	ManufacturerData []Data    `json:"manufacturerData"`
	ServiceData      []Data    `json:"serviceData"`
	RawAdvertisement string    `json:"rawAdvertisement,omitempty"`
	FirstSeen        time.Time `json:"firstSeen"`
	LastSeen         time.Time `json:"lastSeen"`
	SeenCount        int       `json:"seenCount"`
}

type Data struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func Scan(ctx context.Context, opts Options) ([]Device, error) {
	if opts.Duration <= 0 {
		return nil, errors.New("scan duration must be greater than zero")
	}

	adapter := bluetooth.DefaultAdapter
	if strings.TrimSpace(opts.AdapterID) != "" {
		adapter = bluetooth.NewAdapter(opts.AdapterID)
	}

	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("enable BLE stack: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Duration)
	defer cancel()

	var mu sync.Mutex
	devicesByAddress := make(map[string]Device)

	done := make(chan error, 1)
	go func() {
		done <- adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			device := deviceFromScanResult(result, time.Now())

			mu.Lock()
			if existing, ok := devicesByAddress[device.Address]; ok {
				device.FirstSeen = existing.FirstSeen
				device.SeenCount = existing.SeenCount + 1
				if device.Name == "" {
					device.Name = existing.Name
				}
				if len(device.ServiceUUIDs) == 0 {
					device.ServiceUUIDs = existing.ServiceUUIDs
				}
				if len(device.ManufacturerData) == 0 {
					device.ManufacturerData = existing.ManufacturerData
				}
				if len(device.ServiceData) == 0 {
					device.ServiceData = existing.ServiceData
				}
				if device.RawAdvertisement == "" {
					device.RawAdvertisement = existing.RawAdvertisement
				}
			}
			devicesByAddress[device.Address] = device
			mu.Unlock()
		})
	}()

	select {
	case <-ctx.Done():
		if err := adapter.StopScan(); err != nil {
			return nil, fmt.Errorf("stop scan: %w", err)
		}
		if err := <-done; err != nil {
			return nil, fmt.Errorf("scan bluetooth devices: %w", err)
		}
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("scan bluetooth devices: %w", err)
		}
	}

	mu.Lock()
	devices := make([]Device, 0, len(devicesByAddress))
	for _, device := range devicesByAddress {
		devices = append(devices, device)
	}
	mu.Unlock()

	sort.Slice(devices, func(i, j int) bool {
		if devices[i].RSSI == devices[j].RSSI {
			return devices[i].Address < devices[j].Address
		}
		return devices[i].RSSI > devices[j].RSSI
	})

	return devices, nil
}

func deviceFromScanResult(result bluetooth.ScanResult, now time.Time) Device {
	serviceUUIDs := make([]string, 0, len(result.ServiceUUIDs()))
	for _, uuid := range result.ServiceUUIDs() {
		serviceUUIDs = append(serviceUUIDs, uuid.String())
	}
	sort.Strings(serviceUUIDs)

	manufacturerData := make([]Data, 0, len(result.ManufacturerData()))
	for _, item := range result.ManufacturerData() {
		manufacturerData = append(manufacturerData, Data{
			ID:   fmt.Sprintf("0x%04x", item.CompanyID),
			Data: hex.EncodeToString(item.Data),
		})
	}

	serviceData := make([]Data, 0, len(result.ServiceData()))
	for _, item := range result.ServiceData() {
		serviceData = append(serviceData, Data{
			ID:   item.UUID.String(),
			Data: hex.EncodeToString(item.Data),
		})
	}

	addressType := "public"
	if result.Address.IsRandom() {
		addressType = "random"
	}

	rawAdvertisement := ""
	if raw := result.Bytes(); len(raw) > 0 {
		rawAdvertisement = hex.EncodeToString(raw)
	}

	return Device{
		Address:          result.Address.String(),
		AddressType:      addressType,
		RSSI:             result.RSSI,
		Name:             result.LocalName(),
		ServiceUUIDs:     serviceUUIDs,
		ManufacturerData: manufacturerData,
		ServiceData:      serviceData,
		RawAdvertisement: rawAdvertisement,
		FirstSeen:        now,
		LastSeen:         now,
		SeenCount:        1,
	}
}
