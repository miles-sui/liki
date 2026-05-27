package http

type birthParams struct {
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      int     `json:"hour"`
	Minute    int     `json:"minute"`
	Longitude float64 `json:"longitude"`
	Timezone  float64 `json:"timezone"`
	IsDST     bool    `json:"is_dst"`
	Gender    string  `json:"gender"`
}

func validateBirthInfo(bp birthParams) error {
	if bp.Year < 1900 || bp.Year > 2200 {
		return errMsg("year out of range (1900-2200)")
	}
	if bp.Month < 1 || bp.Month > 12 {
		return errMsg("month must be 1-12")
	}
	if bp.Day < 1 || bp.Day > 31 {
		return errMsg("day must be 1-31")
	}
	if bp.Hour < 0 || bp.Hour > 23 {
		return errMsg("hour must be 0-23")
	}
	if bp.Minute < 0 || bp.Minute > 59 {
		return errMsg("minute must be 0-59")
	}
	if bp.Longitude < -180 || bp.Longitude > 180 {
		return errMsg("longitude must be -180 to 180")
	}
	if bp.Gender != "" && bp.Gender != "male" && bp.Gender != "female" {
		return errMsg("gender must be 'male' or 'female'")
	}
	return nil
}

type errMsg string

func (e errMsg) Error() string { return string(e) }
