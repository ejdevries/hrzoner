package netdevices

import (
	"net"
	"reflect"
	"testing"
)

func TestHostsSkipsNetworkAndBroadcast(t *testing.T) {
	_, network, err := net.ParseCIDR("192.168.1.0/30")
	if err != nil {
		t.Fatal(err)
	}

	got, err := hosts(network, 16)
	if err != nil {
		t.Fatal(err)
	}

	var gotStrings []string
	for _, ip := range got {
		gotStrings = append(gotStrings, ip.String())
	}

	want := []string{"192.168.1.1", "192.168.1.2"}
	if !reflect.DeepEqual(gotStrings, want) {
		t.Fatalf("hosts() = %v, want %v", gotStrings, want)
	}
}

func TestHostsRejectsNetworkAboveLimit(t *testing.T) {
	_, network, err := net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	_, err = hosts(network, 4096)
	if err == nil {
		t.Fatal("expected hosts to reject network above max host limit")
	}
}

func TestParsePorts(t *testing.T) {
	got, err := ParsePorts("22, 80,443")
	if err != nil {
		t.Fatal(err)
	}

	want := []int{22, 80, 443}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParsePorts() = %v, want %v", got, want)
	}
}

func TestParsePortsRejectsTrailingText(t *testing.T) {
	_, err := ParsePorts("22abc")
	if err == nil {
		t.Fatal("expected ParsePorts to reject invalid port")
	}
}
