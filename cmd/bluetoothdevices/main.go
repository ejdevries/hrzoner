package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ejdevries/hrzoner/pkg/bluetoothscanner"
	"github.com/ejdevries/hrzoner/pkg/config"
)

func main() {
	durationFlag := flag.Duration("duration", 10*time.Second, "Bluetooth LE scan duration")
	adapterFlag := flag.String("adapter", "", "Bluetooth adapter id, for example hci0; defaults to tinygo's default adapter")
	configFlag := flag.String("config", "config/config.yaml", "config file used to mark the configured HRM device")
	jsonFlag := flag.Bool("json", false, "print full scan results as JSON")
	flag.Parse()

	configuredHRM := strings.TrimSpace(config.ReadConfigPath(*configFlag).HrmConfig.HrmDeviceId)

	devices, err := bluetoothscanner.Scan(context.Background(), bluetoothscanner.Options{
		AdapterID: *adapterFlag,
		Duration:  *durationFlag,
	})
	if err != nil {
		exit(err)
	}

	if *jsonFlag {
		printJSON(devices)
		return
	}

	printTable(devices, configuredHRM)
}

func printTable(devices []bluetoothscanner.Device, configuredHRM string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ADDRESS\tTYPE\tRSSI\tSEEN\tNAME\tSERVICES\tMANUFACTURER_DATA\tSERVICE_DATA\tMATCH")
	for _, device := range devices {
		match := ""
		if strings.EqualFold(device.Address, configuredHRM) {
			match = "configured-hrm"
		}
		fmt.Fprintf(
			writer,
			"%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
			device.Address,
			device.AddressType,
			device.RSSI,
			device.SeenCount,
			valueOrDash(device.Name),
			valueOrDash(strings.Join(device.ServiceUUIDs, ",")),
			valueOrDash(formatData(device.ManufacturerData)),
			valueOrDash(formatData(device.ServiceData)),
			valueOrDash(match),
		)
	}
	writer.Flush()

	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "No Bluetooth LE devices found during the scan window.")
	}
}

func printJSON(devices []bluetoothscanner.Device) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(devices); err != nil {
		exit(err)
	}
}

func formatData(items []bluetoothscanner.Data) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.ID+"="+item.Data)
	}
	return strings.Join(values, ",")
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
