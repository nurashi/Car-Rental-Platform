package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
)

var ErrLicenseNotFound = errors.New("no driver license found")

type LicenseService struct {
	licenseRepo repository.LicenseRepository
}

func NewLicenseService(licenseRepo repository.LicenseRepository) *LicenseService {
	return &LicenseService{licenseRepo: licenseRepo}
}

type SubmitLicenseInput struct {
	UserID         string
	LicenseNumber  string
	IssuingCountry string
	IssueDate      string
	ExpiryDate     string
	FrontImageURL  string
	BackImageURL   string
}

func (s *LicenseService) SubmitLicense(ctx context.Context, input SubmitLicenseInput) (*domain.DriverLicense, error) {
	issueDate, err := time.Parse(time.DateOnly, input.IssueDate)
	if err != nil {
		return nil, fmt.Errorf("parse issue date: %w", err)
	}

	expiryDate, err := time.Parse(time.DateOnly, input.ExpiryDate)
	if err != nil {
		return nil, fmt.Errorf("parse expiry date: %w", err)
	}

	if expiryDate.Before(time.Now()) {
		return nil, errors.New("license has expired")
	}

	license := &domain.DriverLicense{
		UserID:         input.UserID,
		LicenseNumber:  input.LicenseNumber,
		IssuingCountry: input.IssuingCountry,
		IssueDate:      issueDate,
		ExpiryDate:     expiryDate,
		FrontImageURL:  input.FrontImageURL,
		BackImageURL:   input.BackImageURL,
		Status:         domain.LicensePending,
	}

	if err := s.licenseRepo.Create(ctx, license); err != nil {
		return nil, fmt.Errorf("create license: %w", err)
	}

	return license, nil
}

func (s *LicenseService) ValidateLicense(ctx context.Context, userID string) (*domain.DriverLicense, bool, error) {
	license, err := s.licenseRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, false, ErrLicenseNotFound
		}
		return nil, false, fmt.Errorf("get license: %w", err)
	}

	isValid := license.Status == domain.LicenseValid && license.ExpiryDate.After(time.Now())
	return license, isValid, nil
}
