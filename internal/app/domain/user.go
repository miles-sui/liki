package domain

import (
	"encoding/json"
	"strings"
	"time"
)

// Gender drives Big Fortune direction calculation.
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// BirthInfo is the input for a chart computation, stored as JSON on User.
type BirthInfo struct {
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      int     `json:"hour"`
	Minute    int     `json:"minute"`
	Longitude float64 `json:"longitude"`
	Timezone  float64 `json:"timezone"`
	IsDST     bool    `json:"is_dst"`
	Gender    Gender  `json:"gender"`
	Locale    string  `json:"locale"`
}

// User is the central entity for the User aggregate.
type User struct {
	ID              int64
	Name            string
	PasswordHash    string
	TokenVersion    int
	Email           string
	EmailVerifiedAt *time.Time
	PendingEmail    *string
	IsPublic        bool
	DeactivatedAt   *time.Time
	SupporterSince  *time.Time
	BirthInfo       *BirthInfo
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SetBirthInfoJSON stores birth info as a JSON string.
func (u *User) SetBirthInfoJSON(raw string) {
	if raw == "" || raw == "null" {
		u.BirthInfo = nil
		return
	}
	var bi BirthInfo
	if err := json.Unmarshal([]byte(raw), &bi); err == nil {
		u.BirthInfo = &bi
	}
}

// BirthInfoJSON returns the JSON representation of BirthInfo, or empty string if nil.
func (u *User) BirthInfoJSON() string {
	if u.BirthInfo == nil {
		return ""
	}
	b, _ := json.Marshal(u.BirthInfo)
	return string(b)
}

// ValidateRegistration checks name, email, and password constraints.
func ValidateRegistration(name, email, password string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" || password == "" {
		return ErrNameAndPasswordRequired
	}
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if strings.Contains(strings.ToLower(password), strings.ToLower(name)) {
		return ErrPasswordContainsName
	}
	if IsReservedName(name) {
		return ErrUsernameReserved
	}
	return nil
}

// isValidEmail performs basic email format validation.
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	// Must have a dot in the domain part.
	dot := strings.LastIndexByte(email, '.')
	if dot <= at+1 || dot == len(email)-1 {
		return false
	}
	return true
}

// reservedNames are URL paths that cannot be used as usernames.
var reservedNames = map[string]bool{
	"login": true, "register": true, "assess": true, "result": true,
	"types": true, "about": true, "faq": true, "privacy": true,
	"terms": true, "cookies": true, "refund": true,
	"forgot-password": true, "reset-password": true, "verify-email": true,
	"api": true, "content": true, "admin": true, "health": true,
	"healthz": true, "debug": true,
	"r": true, "m": true, "p": true, "profile": true, "settings": true,
	"js": true, "css": true, "img": true, "fonts": true,
	"en": true, "zh-CN": true,
}

// IsReservedName reports whether name is a reserved URL path.
func IsReservedName(name string) bool {
	return reservedNames[strings.ToLower(name)]
}

// CanReactivate returns true if the deactivated account is still within the 7-day grace period.
func (u *User) CanReactivate(now time.Time) bool {
	if u.DeactivatedAt == nil {
		return false
	}
	return now.Sub(*u.DeactivatedAt) <= 7*24*time.Hour
}

func (u *User) IsDeactivated() bool {
	return u.DeactivatedAt != nil && !u.DeactivatedAt.IsZero()
}

// ReactivateIfEligible clears the deactivation if within the grace period.
// Returns ErrInvalidCredentials if the grace period has passed.
func (u *User) ReactivateIfEligible(now time.Time) error {
	if !u.CanReactivate(now) {
		return ErrInvalidCredentials
	}
	u.DeactivatedAt = nil
	u.TokenVersion++
	return nil
}
