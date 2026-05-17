package huelight

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ejdevries/hrzoner/pkg/config"
	"github.com/iDigitalFlame/hue"
)

type HueLight struct {
	bridge *hue.Bridge
	group  *hue.Group
}

func NewHueLight(config config.HueConfig) (*HueLight, error) {
	b, err := hue.Connect(config.IpBridge, config.ApiKey)
	if err != nil {
		os.Stderr.WriteString(`Failed to connect to "` + config.IpBridge + `": ` + err.Error() + "!\n")
		os.Exit(1)
	}
	g := b.GroupByName(config.Room)
	hueLight := &HueLight{
		bridge: b,
		group:  g,
	}
	return hueLight, nil
}

func (hl *HueLight) SetHexColor(hexColor string) {
	for _, e := range hl.group.Lights {
		err := e.On()
		if err != nil {
			fmt.Println("Failed to turn on light:", err)
			return
		}
		//e.SetOn(true)
		if err := e.SetHex(hexColor); err != nil {
			break
		}
		if err := e.Update(); err != nil {
			break
		}
	}
}

func cycleColor(b *hue.Bridge, addr string, target string) {
	fmt.Println("cycleColor")
	i := b.GroupByName(target)
	if i == nil {
		os.Stderr.WriteString(`Could not find target "` + target + `" on "` + addr + `"!` + "\n")
		os.Exit(1)
	}
	wasOn := i.On
	fmt.Println(wasOn)
	for c := 0x000000; c < 0xFFFFFF; c = c + 0x000040 {
		var err error
		fmt.Println(c)
		hexString := fmt.Sprintf("#%06X", c)
		fmt.Println(hexString)
		time.Sleep(500 * time.Millisecond)
		for _, e := range i.Lights {
			e.On()
			//e.SetOn(true)
			if err = e.SetHex(hexString); err != nil {
				break
			}
			if err = e.Update(); err != nil {
				break
			}
		}
	}
}

func cycleHrColor(b *hue.Bridge, addr string, target string) {
	fmt.Println("cycleColor")
	i := b.GroupByName(target)
	if i == nil {
		os.Stderr.WriteString(`Could not find target "` + target + `" on "` + addr + `"!` + "\n")
		os.Exit(1)
	}
	fmt.Println(i.On)
	hrColors := []string{"#CCCCCC", "#AAAAAA", "#00AAFF", "#00FF00", "#FFAA00", "#FF0000", "#CCCCCC", "#777777", "#444444", "#222222"}

	for _, c := range hrColors {
		var err error
		fmt.Println(c)
		time.Sleep(1200 * time.Millisecond)
		for _, e := range i.Lights {
			e.On()
			//e.SetOn(true)
			if err = e.SetHex(c); err != nil {
				break
			}
			if err = e.Update(); err != nil {
				break
			}
		}
	}
}

func list(br *hue.Bridge, addr string) {

	g, err := br.Groups()
	if err != nil {
		os.Stderr.WriteString(`Could not get the Groups list from "` + addr + `": ` + err.Error() + "!\n")
		os.Exit(1)
	}
	os.Stdout.WriteString("Groups List\n================\n")
	for k, v := range g {
		os.Stdout.WriteString(
			"[" + k + "] " + v.Name() + ": " + v.Type.String() + " Lights " +
				strconv.FormatUint(uint64(len(v.Lights)), 10) + "\n",
		)
	}
	l, err := br.Lights()
	if err != nil {
		os.Stderr.WriteString(`Could not get the Lights list from "` + addr + `": ` + err.Error() + "!\n")
		os.Exit(1)
	}
	os.Stdout.WriteString("\nLights List\n================\n")
	for k, v := range l {
		if os.Stdout.WriteString("[" + k + "] " + v.Name() + ": " + v.Model + " "); v.IsOn() {
			os.Stdout.WriteString("On Brightness " + strconv.FormatUint(uint64((float32(v.Brightness())/254.0)*100.0), 10) + "% ")
			if v.IsColor() {
				os.Stdout.WriteString(
					"Hue: " + strconv.FormatUint(uint64(v.Hue()), 10) +
						" Sat: " + strconv.FormatUint(uint64(v.Saturation()), 10) +
						" Temp: " + strconv.FormatUint(uint64(v.Temperature()), 10),
				)
			}
			os.Stdout.WriteString("\n")
			continue
		}
		os.Stdout.WriteString("Off\n")
	}
}

func testLight() {

	url := "http://192.168.2.5/api/7P4dB3itlnSrQkU1EN80PjBwHVR0GuIirDCbX4Us/lights/13/state"
	method := "PUT"

	payload := strings.NewReader(` 	{"on":true, "sat":254, "bri":254,"hue":65432}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
