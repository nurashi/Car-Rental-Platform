package pb

import (
	"time"

	"github.com/nurashi/car-rental-identity/internal/domain"
)

func UserToProto(u domain.User) *User {
	return &User{
		Id:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
		Status:        string(u.Status),
		AvatarUrl:     u.AvatarURL,
		CreatedAt:     u.CreatedAt.Unix(),
		UpdatedAt:     u.UpdatedAt.Unix(),
	}
}

func UsersToProto(users []domain.User) []*User {
	result := make([]*User, len(users))
	for i, u := range users {
		result[i] = UserToProto(u)
	}
	return result
}

func DriverLicenseToProto(l domain.DriverLicense) *DriverLicense {
	return &DriverLicense{
		Id:             l.ID,
		UserId:         l.UserID,
		LicenseNumber:  l.LicenseNumber,
		IssuingCountry: l.IssuingCountry,
		IssueDate:      l.IssueDate.Format(time.DateOnly),
		ExpiryDate:     l.ExpiryDate.Format(time.DateOnly),
		Status:         string(l.Status),
		CreatedAt:      l.CreatedAt.Unix(),
		UpdatedAt:      l.UpdatedAt.Unix(),
	}
}

func NotificationToProto(n domain.NotificationRecord) *Notification {
	return &Notification{
		Id:        n.ID,
		UserId:    n.UserID,
		Type:      n.Type,
		Subject:   n.Subject,
		Body:      n.Body,
		CreatedAt: n.CreatedAt.Unix(),
	}
}

func NotificationsToProto(records []domain.NotificationRecord) []*Notification {
	result := make([]*Notification, len(records))
	for i, n := range records {
		result[i] = NotificationToProto(n)
	}
	return result
}
