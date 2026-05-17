package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ejdevries/hrzoner/pkg/netdevices"
)

func main() {
	cidrFlag := flag.String("cidr", "", "comma-separated IPv4 CIDRs to scan; defaults to active local networks")
	portsFlag := flag.String("ports", netdevices.FormatPorts(netdevices.DefaultPorts), "comma-separated TCP ports to probe")
	timeoutFlag := flag.Duration("timeout", 250*time.Millisecond, "timeout per TCP connection attempt")
	concurrencyFlag := flag.Int("concurrency", 128, "maximum concurrent host probes")
	maxHostsFlag := flag.Int("max-hosts", 4096, "maximum addresses allowed per CIDR")
	flag.Parse()

	cidrs, err := netdevices.ParseCIDRs(*cidrFlag)
	if err != nil {
		exit(err)
	}

	ports, err := netdevices.ParsePorts(*portsFlag)
	if err != nil {
		exit(err)
	}

	opts := netdevices.Options{
		CIDRs:       cidrs,
		Ports:       ports,
		Timeout:     *timeoutFlag,
		Concurrency: *concurrencyFlag,
		MaxHosts:    *maxHostsFlag,
	}

	devices, err := netdevices.Discover(context.Background(), opts)
	if err != nil {
		exit(err)
	}

	fmt.Println("IP\tHOSTNAME\tOPEN_PORTS")
	for _, device := range devices {
		hostname := device.Hostname
		if strings.TrimSpace(hostname) == "" {
			hostname = "-"
		}
		fmt.Printf("%s\t%s\t%s\n", device.IP, hostname, netdevices.FormatPorts(device.OpenPorts))
	}

	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "No devices found. Hosts with no reachable probed TCP ports are not reported.")
	}
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
