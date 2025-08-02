package authconstants

const (
	ErrInvalidCredentials      = "invalid email or password"
	ErrAccountNotFound         = "account not found"
	ErrAccountAlreadyExists    = "account with this email already exists"
	ErrEmailNotVerified        = "email not verified"
	ErrPhoneNotVerified        = "phone not verified"
	ErrInvalidToken            = "invalid token"
	ErrTokenExpired            = "token expired"
	ErrTokenAlreadyUsed        = "token already used"
	ErrSessionExpired          = "session expired"
	ErrInvalidPassword         = "invalid password"
	ErrPasswordTooWeak         = "password too weak"
	ErrInvalidEmail            = "invalid email format"
	ErrInvalidPhone            = "invalid phone format"
	ErrVerificationRequired    = "verification required"
	ErrUnauthorized            = "unauthorized"
)