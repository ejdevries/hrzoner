package netdevices

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// var DefaultPorts = []int{22, 53, 80, 135, 139, 443, 445, 548, 554, 631, 8080, 8443}
var DefaultPorts = []int{80, 443, 8080, 8443}

type Device struct {
	IP        net.IP
	Hostname  string
	OpenPorts []int
}

type Options struct {
	CIDRs       []*net.IPNet
	Ports       []int
	Timeout     time.Duration
	Concurrency int
	MaxHosts    int
}

func LocalIPv4Networks() ([]*net.IPNet, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var networks []*net.IPNet
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			ip, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil {
				continue
			}
			ipNet.IP = ip.To4().Mask(ipNet.Mask)

			ones, bits := ipNet.Mask.Size()
			if bits == 32 && ones < 32 {
				networks = append(networks, ipNet)
			}
		}
	}

	if len(networks) == 0 {
		return nil, errors.New("no active non-loopback IPv4 networks found")
	}
	return networks, nil
}

func Discover(ctx context.Context, opts Options) ([]Device, error) {
	opts = normalizeOptions(opts)

	if len(opts.CIDRs) == 0 {
		networks, err := LocalIPv4Networks()
		if err != nil {
			return nil, err
		}
		opts.CIDRs = networksWithinLimit(networks, opts.MaxHosts)
		if len(opts.CIDRs) == 0 {
			return nil, fmt.Errorf("no local IPv4 networks within max-hosts %d", opts.MaxHosts)
		}
	}

	ips, err := scanTargets(opts.CIDRs, opts.MaxHosts)
	if err != nil {
		return nil, err
	}

	jobs := make(chan net.IP)
	results := make(chan Device)

	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				device := probe(ctx, ip, opts.Ports, opts.Timeout)
				if len(device.OpenPorts) > 0 {
					results <- device
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				return
			case jobs <- ip:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var devices []Device
	for device := range results {
		devices = append(devices, device)
	}
	if err := ctx.Err(); err != nil {
		return devices, err
	}

	sort.Slice(devices, func(i, j int) bool {
		return ipToUint32(devices[i].IP) < ipToUint32(devices[j].IP)
	})

	return devices, nil
}

func ParseCIDRs(value string) ([]*net.IPNet, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	var cidrs []*net.IPNet
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ip, ipNet, err := net.ParseCIDR(part)
		if err != nil {
			return nil, fmt.Errorf("parse CIDR %q: %w", part, err)
		}
		if ip.To4() == nil {
			return nil, fmt.Errorf("CIDR %q is not an IPv4 network", part)
		}

		ipNet.IP = ip.To4().Mask(ipNet.Mask)
		cidrs = append(cidrs, ipNet)
	}

	return cidrs, nil
}

func ParsePorts(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	var ports []int
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		port, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("parse port %q: %w", part, err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port %d is outside 1-65535", port)
		}
		ports = append(ports, port)
	}

	return ports, nil
}

func FormatPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}

	values := make([]string, 0, len(ports))
	for _, port := range ports {
		values = append(values, fmt.Sprintf("%d", port))
	}
	return strings.Join(values, ",")
}

func normalizeOptions(opts Options) Options {
	if len(opts.Ports) == 0 {
		opts.Ports = DefaultPorts
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 250 * time.Millisecond
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 128
	}
	if opts.MaxHosts <= 0 {
		opts.MaxHosts = 4096
	}
	return opts
}

func scanTargets(cidrs []*net.IPNet, maxHosts int) ([]net.IP, error) {
	seen := make(map[uint32]struct{})
	var ips []net.IP

	for _, cidr := range cidrs {
		hosts, err := hosts(cidr, maxHosts)
		if err != nil {
			return nil, err
		}
		for _, ip := range hosts {
			key := ipToUint32(ip)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ips = append(ips, ip)
		}
	}

	sort.Slice(ips, func(i, j int) bool {
		return ipToUint32(ips[i]) < ipToUint32(ips[j])
	})

	return ips, nil
}

func networksWithinLimit(networks []*net.IPNet, maxHosts int) []*net.IPNet {
	var filtered []*net.IPNet
	for _, network := range networks {
		count, ok := addressCount(network)
		if !ok {
			continue
		}
		if maxHosts <= 0 || count <= uint64(maxHosts) {
			filtered = append(filtered, network)
		}
	}
	return filtered
}

func hosts(network *net.IPNet, maxHosts int) ([]net.IP, error) {
	_, bits := network.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("%s is not an IPv4 network", network.String())
	}

	base := ipToUint32(network.IP.To4()) & maskToUint32(network.Mask)
	count, _ := addressCount(network)
	if maxHosts > 0 && count > uint64(maxHosts) {
		return nil, fmt.Errorf("%s has %d addresses, above max-hosts %d", network.String(), count, maxHosts)
	}

	start := base
	end := base + uint32(count) - 1
	if count > 2 {
		start++
		end--
	}

	ips := make([]net.IP, 0, end-start+1)
	for current := start; current <= end; current++ {
		ips = append(ips, uint32ToIP(current))
		if current == end {
			break
		}
	}
	return ips, nil
}

func addressCount(network *net.IPNet) (uint64, bool) {
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return 0, false
	}
	return uint64(1) << uint(bits-ones), true
}

func probe(ctx context.Context, ip net.IP, ports []int, timeout time.Duration) Device {
	device := Device{IP: ip}

	for _, port := range ports {
		address := net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
		dialer := net.Dialer{Timeout: timeout}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			continue
		}
		conn.Close()
		device.OpenPorts = append(device.OpenPorts, port)
	}

	if len(device.OpenPorts) == 0 {
		return device
	}

	lookupCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(lookupCtx, ip.String())
	if err == nil && len(names) > 0 {
		device.Hostname = strings.TrimSuffix(names[0], ".")
	}

	return device
}

func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func uint32ToIP(value uint32) net.IP {
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, value)
	return ip
}

func maskToUint32(mask net.IPMask) uint32 {
	return binary.BigEndian.Uint32(mask)
}
