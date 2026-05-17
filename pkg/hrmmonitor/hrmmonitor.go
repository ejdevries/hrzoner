package hrmmonitor

import (
	"time"

	"tinygo.org/x/bluetooth"
)

var (
	adapter                     = bluetooth.DefaultAdapter
	heartRateServiceUUID        = bluetooth.ServiceUUIDHeartRate
	heartRateCharacteristicUUID = bluetooth.CharacteristicUUIDHeartRateMeasurement
)

type HrmMonitor struct {
	device  *bluetooth.Device
	service *bluetooth.DeviceService
	Char    *bluetooth.DeviceCharacteristic
}

func NewHrmMonitor(hrmDeviceId string) (*HrmMonitor, error) {
	println("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	ch := make(chan bluetooth.ScanResult, 1)
	// EE:EC:ED:1D:85:B9 -63 HRM-Dual:715001
	// Start scanning.
	println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == hrmDeviceId {
			err := adapter.StopScan()
			if err != nil {
				return
			}
			ch <- result
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})

		if err != nil {
			println(err.Error())
			return nil, err
		}

		println("connected to ", result.Address.String())
	}

	// get services
	println("discovering services/characteristics")
	services, err := device.DiscoverServices([]bluetooth.UUID{heartRateServiceUUID})
	must("discover services", err)

	if len(services) == 0 {
		panic("could not find heart rate service")
	}

	service := services[0]

	println("found service", service.UUID().String())

	chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{heartRateCharacteristicUUID})
	if err != nil {
		println(err)
		return nil, err
	}

	if len(chars) == 0 {
		panic("could not find heart rate characteristic")
	}

	char := chars[0]
	println("found characteristic", char.UUID().String())

	//char.EnableNotifications(func(buf []byte) {
	//	//hr := int(buf[1])
	//	println("data:", uint8(buf[1]))
	//	//println(hz.GetZoneColor(hr))
	//	println(time.Now().String())
	//})

	hrmMonitor := &HrmMonitor{
		device:  &device,
		service: &service,
		Char:    &char,
	}
	return hrmMonitor, nil
}

func (hrm *HrmMonitor) Notify(callback func(buf []byte)) {
	println("start notify")
	err := hrm.Char.EnableNotifications(callback)
	//err := hrm.char.EnableNotifications(func(buf []byte) {
	//	//hr := int(buf[1])
	//	println("data:", uint8(buf[1]))
	//	//println(hz.GetZoneColor(hr))
	//	println(time.Now().String())
	//})
	if err != nil {
		return
	}
}

func main() {
	println("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())
	connectAddress := "EE:EC:ED:1D:85:B9"

	ch := make(chan bluetooth.ScanResult, 1)
	// EE:EC:ED:1D:85:B9 -63 HRM-Dual:715001
	// Start scanning.
	println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == connectAddress {
			adapter.StopScan()
			ch <- result
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			println(err.Error())
			return
		}

		println("connected to ", result.Address.String())
	}

	// get services
	println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices([]bluetooth.UUID{heartRateServiceUUID})
	must("discover services", err)

	if len(srvcs) == 0 {
		panic("could not find heart rate service")
	}

	srvc := srvcs[0]

	println("found service", srvc.UUID().String())

	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{heartRateCharacteristicUUID})
	if err != nil {
		println(err)
	}

	if len(chars) == 0 {
		panic("could not find heart rate characteristic")
	}

	char := chars[0]
	println("found characteristic", char.UUID().String())

	//hz := NewHrZones(180)

	char.EnableNotifications(func(buf []byte) {
		//hr := int(buf[1])
		println("data:", uint8(buf[1]))
		//println(hz.GetZoneColor(hr))
		println(time.Now().String())
	})

}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
