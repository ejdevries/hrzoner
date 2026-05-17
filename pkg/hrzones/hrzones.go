package hrzones

import "math"

const hrColorLtGray = "#AAAAAA"
const hrColorBlue = "#00AAFF"
const hrColorGreen = "#00FF00"
const hrColorYellow = "#FFAA00"
const hrColorRed = "#FF0000"

type hrZone struct {
	minHr     int
	zoneColor string
}

type HrZones struct {
	maxHr int
	zones []hrZone
}

func NewHrZones(maxHrUser int) HrZones {
	hrValues := []float64{0.0, 0.5, 0.6, 0.7, 0.8, 0.9}
	hrColors := []string{hrColorLtGray, hrColorLtGray, hrColorBlue, hrColorGreen, hrColorYellow, hrColorRed}

	var hz []hrZone
	for i, v := range hrValues {
		//println(i, i2)
		hz = append(hz, hrZone{minHr: int(math.Round(v * float64(maxHrUser))), zoneColor: hrColors[i]})
	}
	return HrZones{maxHr: maxHrUser, zones: hz}
}

func (hz *HrZones) GetZoneColor(hr int) string {
	zoneColor := hrColorLtGray
	for _, v := range hz.zones {
		if hr >= v.minHr {
			zoneColor = v.zoneColor
			continue
		}
	}
	return zoneColor
}

//        heartRateZoneColors = [
//   Gfx.COLOR_LT_GRAY, 0xAAAAAA z1 - 50 - 60
//   Gfx.COLOR_BLUE, 0x00AAFF z2 - 60 - 70
//   Gfx.COLOR_GREEN, 0x00FF00 z3 - 70 - 80
//   Gfx.COLOR_YELLOW, 0xFFAA00 - z4 - 80 - 90
//   Gfx.COLOR_RED] 0xFF0000 - Z5 - 90 - 100%
//for (var i = 0; i < heartRateZones.size(); i++) {
//heartRateZoneColor = heartRateZoneColors[i];
//if (currentHeartRate <= heartRateZones[i]) {
//break;
//}
//}

// max HR 180
// resting heart rate 46
// heart rate reserve 134
//Your Running Heart Rate Zone Summary:
//Running ZONE 1 calculator / Z1< 126
//Running ZONE 2 calculator / Z2127 - 140
//Running ZONE 3 calculator / Z3141 - 153
//Running ZONE 4 calculator / Z4154 - 167
//Running ZONE 5 calculator / Z5168 - 180

//erik zones Strava 0-109 109-144 144-161 161-179 > 179
//erik zones Garmin 95-112 113-131 132-150 151-169 > 170 - 189 (en verder) - edge zegt max hr 189 lactate treshold 169
