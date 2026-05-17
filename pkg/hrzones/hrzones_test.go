package hrzones

import (
	"reflect"
	"testing"
)

func TestNewHrZones(t *testing.T) {
	type args struct {
		maxHrUser int
	}
	tests := []struct {
		name string
		args args
		want HrZones
	}{
		{name: "test max hr 100", args: args{maxHrUser: 100}, want: HrZones{maxHr: 100, zones: []hrZone{{minHr: 0, zoneColor: "#AAAAAA"}, {50, "#AAAAAA"}, {60, "#00AAFF"}, {70, "#00FF00"}, {80, "#FFAA00"}, {90, "#FF0000"}}}},
		{name: "test max hr 180", args: args{maxHrUser: 180}, want: HrZones{maxHr: 180, zones: []hrZone{{minHr: 0, zoneColor: "#AAAAAA"}, {90, "#AAAAAA"}, {108, "#00AAFF"}, {126, "#00FF00"}, {144, "#FFAA00"}, {162, "#FF0000"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHrZones(tt.args.maxHrUser); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHrZones() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetZoneColor(t *testing.T) {
	maxHr := 180
	hz := NewHrZones(maxHr)

	tests := []struct {
		name string
		args int
		want string
	}{
		{name: "test hr 0", args: 0, want: hrColorLtGray},
		{name: "test hr 180", args: 180, want: hrColorRed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hz.GetZoneColor(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetZoneColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
