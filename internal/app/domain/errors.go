package domain

import "errors"

// -- User aggregate errors --

var (
	ErrNameAndPasswordRequired = errors.New("name and password are required")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters")
	ErrPasswordContainsName    = errors.New("password must not contain username")
	ErrUsernameTaken           = errors.New("username already exists")
	ErrEmailTaken              = errors.New("email already in use")
	ErrInvalidCredentials      = errors.New("invalid username or password")
	ErrCurrentPasswordWrong    = errors.New("current password is incorrect")
	ErrTokenExpired            = errors.New("token expired")
	ErrUserNotFound            = errors.New("user not found")
	ErrNameEmpty               = errors.New("name cannot be empty")
	ErrNoFields                = errors.New("at least one field is required")
	ErrUsernameReserved        = errors.New("username is reserved")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrEmailAlreadyVerified    = errors.New("email already verified by another user")
	ErrNoEmailToVerify         = errors.New("no email to verify")
)

// -- Assessment aggregate errors --

var (
	ErrNoProfile       = errors.New("no profile found — submit an assessment first")
	ErrAnswersRequired = errors.New("answers is required")
)

// -- ReviewLink aggregate errors --

var (
	ErrLinkNotFound = errors.New("review link not found")
	ErrLinkExpired  = errors.New("review link has expired")
)

// -- Match Link errors --

var (
	ErrMatchLinkNotFound = errors.New("match link not found or deleted")
)

// -- Mingli errors --

var (
	ErrInvalidBirthInfo  = errors.New("invalid birth info")
	ErrYearOutOfRange    = errors.New("year out of range (1900-2200)")
	ErrChartRequired     = errors.New("chart a or b is required")
	ErrCityNotFound      = errors.New("city not found")
)

// -- Report errors --

var (
	ErrReportNotFound = errors.New("report not found")
	ErrInvalidScene   = errors.New("invalid scene")
	ErrEngineDataReq  = errors.New("engine_data is required")
)
