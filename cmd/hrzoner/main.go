package main

import (
	"fmt"
	"time"

	"github.com/ejdevries/hrzoner/pkg/config"
	"github.com/ejdevries/hrzoner/pkg/hrmmonitor"
	"github.com/ejdevries/hrzoner/pkg/hrzones"
	"github.com/ejdevries/hrzoner/pkg/huelight"
)

func main() {
	fmt.Println("Starting...")

	appConfig := config.ReadConfig()
	hrZones := hrzones.NewHrZones(appConfig.HrmConfig.MaxHr)
	hueLight, err := huelight.NewHueLight(appConfig.HueConfig)
	if err != nil {
		println(err.Error())
	}
	hexColor := hrZones.GetZoneColor(140)
	hueLight.SetHexColor(hexColor)

	fmt.Printf("%v\n", appConfig)
	fmt.Printf("%v\n", hrZones)
	fmt.Printf("%v\n", hueLight)

	hrmMonitor, err := hrmmonitor.NewHrmMonitor(appConfig.HrmConfig.HrmDeviceId)

	err = hrmMonitor.Char.EnableNotifications(func(buf []byte) {
		hr := int(buf[1])
		println("data:", uint8(buf[1]))
		println(hrZones.GetZoneColor(hr))
		hexColor := hrZones.GetZoneColor(int(buf[1]))
		hueLight.SetHexColor(hexColor)
		println(time.Now().String())
	})
	if err != nil {
		return
	}
	//err := hrm.char.EnableNotifications(func(buf []byte) {
	//	//hr := int(buf[1])
	//	println("data:", uint8(buf[1]))
	//	//println(hz.GetZoneColor(hr))
	//	println(time.Now().String())
	//})

	select {}

	fmt.Println("Finished...")

}
