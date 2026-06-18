package qimen

import (

	"liki/internal/engine/ganzhi"
)

// computePan builds a pan from bureau info and driving pillar.
func computePan(ju juShu, driveZhu ganzhi.Zhu) pan {
	dipan := placeDiPan(ju.Number, ju.YinDun)
	duty := findDuty(driveZhu, dipan)
	tianStars, tianStems := placeTianPan(driveZhu, duty.Star, dipan)
	renDoors := placeRenPan(driveZhu.Zhi, duty.Door)

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
	angans := placeAnGan(driveZhu, dutyDoorPalace)

	mata := findMaXing(driveZhu.Zhi)
	kongWang := findKongWang(driveZhu)

	pan := pan{
		Jushu:    ju.Number,
		YinDun:   ju.YinDun,
		DutyStar: duty.Star,
		DutyDoor: duty.Door,
		MaXing:   mata,
		DriveGan:  driveZhu.Gan,
		DriveZhi: driveZhu.Zhi,
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
