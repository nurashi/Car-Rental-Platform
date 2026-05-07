package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"github.com/nurashi/car-rental-identity/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLicenseRepo struct {
	mock.Mock
}

func (m *mockLicenseRepo) Create(ctx context.Context, license *domain.DriverLicense) error {
	args := m.Called(ctx, license)
	return args.Error(0)
}

func (m *mockLicenseRepo) GetByUserID(ctx context.Context, userID string) (*domain.DriverLicense, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DriverLicense), args.Error(1)
}

func (m *mockLicenseRepo) Update(ctx context.Context, license *domain.DriverLicense) error {
	args := m.Called(ctx, license)
	return args.Error(0)
}

func TestLicenseService_SubmitLicense_Expired(t *testing.T) {
	licenseRepo := new(mockLicenseRepo)
	licenseService := service.NewLicenseService(licenseRepo)

	input := service.SubmitLicenseInput{
		UserID:         "1",
		LicenseNumber:  "ABC123",
		IssuingCountry: "US",
		IssueDate:      "2020-01-01",
		ExpiryDate:     "2021-01-01",
	}

	license, err := licenseService.SubmitLicense(context.Background(), input)

	assert.Nil(t, license)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestLicenseService_SubmitLicense_Success(t *testing.T) {
	licenseRepo := new(mockLicenseRepo)
	licenseRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.DriverLicense")).Return(nil)

	licenseService := service.NewLicenseService(licenseRepo)

	input := service.SubmitLicenseInput{
		UserID:         "1",
		LicenseNumber:  "ABC123",
		IssuingCountry: "KZ",
		IssueDate:      "2024-01-01",
		ExpiryDate:     "2034-01-01",
	}

	license, err := licenseService.SubmitLicense(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, domain.LicensePending, license.Status)
	licenseRepo.AssertExpectations(t)
}

func TestLicenseService_ValidateLicense_Valid(t *testing.T) {
	licenseRepo := new(mockLicenseRepo)
	validLicense := &domain.DriverLicense{
		ID:             "1",
		UserID:         "1",
		Status:         domain.LicenseValid,
		IssueDate:      mustParseTime("2024-01-01"),
		ExpiryDate:     mustParseTime("2034-01-01"),
	}
	licenseRepo.On("GetByUserID", mock.Anything, "1").Return(validLicense, nil)

	licenseService := service.NewLicenseService(licenseRepo)
	license, isValid, err := licenseService.ValidateLicense(context.Background(), "1")

	assert.NoError(t, err)
	assert.True(t, isValid)
	assert.Equal(t, domain.LicenseValid, license.Status)
	licenseRepo.AssertExpectations(t)
}

func TestLicenseService_ValidateLicense_NotFound(t *testing.T) {
	licenseRepo := new(mockLicenseRepo)
	licenseRepo.On("GetByUserID", mock.Anything, "1").Return((*domain.DriverLicense)(nil), repository.ErrLicenseNotFound)

	licenseService := service.NewLicenseService(licenseRepo)
	license, isValid, err := licenseService.ValidateLicense(context.Background(), "1")

	assert.Nil(t, license)
	assert.False(t, isValid)
	assert.Equal(t, service.ErrLicenseNotFound, err)
}

func mustParseTime(s string) time.Time {
	t, _ := time.Parse(time.DateOnly, s)
	return t
}
