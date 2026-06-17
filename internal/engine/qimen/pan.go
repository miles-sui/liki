package qimen

import (

	"liki/internal/engine/ganzhi"
)

// computePan builds a pan from bureau info and driving stem/branch.
func computePan(ju juShu, driveGan ganzhi.Gan, driveZhi ganzhi.Zhi) pan {
	dipan := placeDiPan(ju.Number, ju.YinDun)
	duty := findDuty(driveGan, driveZhi, dipan)
	tianStars, tianStems := placeTianPan(driveGan, duty.Star, dipan)
	renDoors := placeRenPan(driveZhi, duty.Door)

	var dutyStarPalace int
	for i, s := range tianStars {
		if s == duty.Star {
			dutyStarPalace = i
			break
		}
	}
	shenSpirits := placeShenPan(ju.YinDun, PalaceIndex(dutyStarPalace+1))

	var dutyDoorPalace int
	for i, d := range renDoors {
		if d == duty.Door {
			dutyDoorPalace = i
			break
		}
	}
	angans := placeAnGan(driveGan, dutyDoorPalace)

	mata := findMaXing(driveZhi)
	kongWang := findKongWang(driveGan, driveZhi)

	pan := pan{
		Jushu:    ju.Number,
		YinDun:   ju.YinDun,
		DutyStar: duty.Star,
		DutyDoor: duty.Door,
		MaXing:   mata,
		DriveZhi: driveZhi,
		KongWang: kongWang,
	}
	for i := 0; i < 9; i++ {
		pan.Palaces[i] = Palace{
			EarthStem:  dipan[i],
			HeavenStem: tianStems[i],
			Star:       tianStars[i],
			Door:       renDoors[i],
			Spirit:     shenSpirits[i],
			HiddenStem: angans[i],
		}
	}
	return pan
}
