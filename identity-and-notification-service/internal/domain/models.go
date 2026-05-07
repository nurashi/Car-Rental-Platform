package domain

import "time"

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusBlocked  UserStatus = "blocked"
	UserStatusPending  UserStatus = "pending"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	AvatarURL    string
	EmailVerified bool
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DriverLicenseStatus string

const (
	LicensePending  DriverLicenseStatus = "pending"
	LicenseValid    DriverLicenseStatus = "valid"
	LicenseRejected DriverLicenseStatus = "rejected"
	LicenseExpired  DriverLicenseStatus = "expired"
)

type DriverLicense struct {
	ID            string
	UserID        string
	LicenseNumber string
	IssuingCountry string
	IssueDate     time.Time
	ExpiryDate    time.Time
	FrontImageURL string
	BackImageURL  string
	Status        DriverLicenseStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EmailVerification struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type NotificationRecord struct {
	ID        string
	UserID    string
	Type      string
	Subject   string
	Body      string
	CreatedAt time.Time
}
