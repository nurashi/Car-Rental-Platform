package service_test

import (
	"context"
	"testing"

	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockVehicleRepo struct {
	mock.Mock
}

func (m *mockVehicleRepo) CreateVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *mockVehicleRepo) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vehicle), args.Error(1)
}

func (m *mockVehicleRepo) UpdateVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *mockVehicleRepo) DeleteVehicle(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockVehicleRepo) ListVehicles(ctx context.Context, limit, offset int) ([]*domain.Vehicle, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Vehicle), args.Int(1), args.Error(2)
}

func (m *mockVehicleRepo) UpdateVehicleStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockVehicleRepo) CreateLocation(ctx context.Context, location *domain.Location) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *mockVehicleRepo) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Location), args.Error(1)
}

func (m *mockVehicleRepo) ListLocations(ctx context.Context, limit, offset int) ([]*domain.Location, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Location), args.Int(1), args.Error(2)
}

func (m *mockVehicleRepo) CreateMaintenanceRecord(ctx context.Context, record *domain.MaintenanceRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *mockVehicleRepo) GetMaintenanceRecord(ctx context.Context, id string) (*domain.MaintenanceRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MaintenanceRecord), args.Error(1)
}

func (m *mockVehicleRepo) ListMaintenanceRecords(ctx context.Context, vehicleID string, limit, offset int) ([]*domain.MaintenanceRecord, int, error) {
	args := m.Called(ctx, vehicleID, limit, offset)
	return args.Get(0).([]*domain.MaintenanceRecord), args.Int(1), args.Error(2)
}

type mockCacheRepo struct {
	mock.Mock
}

func (m *mockCacheRepo) SetVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	return m.Called(ctx, vehicle).Error(0)
}

func (m *mockCacheRepo) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vehicle), args.Error(1)
}

func (m *mockCacheRepo) DeleteVehicle(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockCacheRepo) SetLocation(ctx context.Context, location *domain.Location) error {
	return m.Called(ctx, location).Error(0)
}

func (m *mockCacheRepo) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Location), args.Error(1)
}

func (m *mockCacheRepo) DeleteLocation(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func TestVehicleService_CreateVehicle(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	vehicle := &domain.Vehicle{
		Make:         "Toyota",
		Model:        "Camry",
		Year:         2024,
		LicensePlate: "ABC-1234",
		Status:       "available",
	}

	repo.On("CreateVehicle", mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)
	cache.On("SetVehicle", mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	created, err := svc.CreateVehicle(context.Background(), vehicle)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Toyota", created.Make)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_GetVehicle_CacheHit(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	cached := &domain.Vehicle{ID: "1", Make: "Toyota", Model: "Camry"}
	cache.On("GetVehicle", mock.Anything, "1").Return(cached, nil)

	svc := service.NewVehicleService(repo, cache, nil)
	result, err := svc.GetVehicle(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, cached, result)
	repo.AssertNotCalled(t, "GetVehicle", mock.Anything, "1")
	cache.AssertExpectations(t)
}

func TestVehicleService_GetVehicle_CacheMiss(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	vehicle := &domain.Vehicle{ID: "1", Make: "Toyota"}
	cache.On("GetVehicle", mock.Anything, "1").Return((*domain.Vehicle)(nil), nil)
	repo.On("GetVehicle", mock.Anything, "1").Return(vehicle, nil)
	cache.On("SetVehicle", mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	result, err := svc.GetVehicle(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, "Toyota", result.Make)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_UpdateVehicle(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	vehicle := &domain.Vehicle{ID: "1", Make: "Toyota", Model: "Camry"}
	repo.On("UpdateVehicle", mock.Anything, vehicle).Return(nil)
	cache.On("SetVehicle", mock.Anything, vehicle).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	updated, err := svc.UpdateVehicle(context.Background(), vehicle)

	assert.NoError(t, err)
	assert.Equal(t, "Camry", updated.Model)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_DeleteVehicle(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	repo.On("DeleteVehicle", mock.Anything, "1").Return(nil)
	cache.On("DeleteVehicle", mock.Anything, "1").Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	err := svc.DeleteVehicle(context.Background(), "1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_ListVehicles(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	vehicles := []*domain.Vehicle{{ID: "1", Make: "Toyota"}}
	repo.On("ListVehicles", mock.Anything, 10, 0).Return(vehicles, 1, nil)

	svc := service.NewVehicleService(repo, cache, nil)
	result, total, err := svc.ListVehicles(context.Background(), 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
	repo.AssertExpectations(t)
}

func TestVehicleService_UpdateVehicleStatus(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	vehicle := &domain.Vehicle{ID: "1", Make: "Toyota", Status: "under maintenance"}
	repo.On("UpdateVehicleStatus", mock.Anything, "1", "under maintenance").Return(nil)
	cache.On("GetVehicle", mock.Anything, "1").Return((*domain.Vehicle)(nil), nil)
	repo.On("GetVehicle", mock.Anything, "1").Return(vehicle, nil)
	cache.On("SetVehicle", mock.Anything, vehicle).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	updated, err := svc.UpdateVehicleStatus(context.Background(), "1", "under maintenance")

	assert.NoError(t, err)
	assert.Equal(t, "under maintenance", updated.Status)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_CreateLocation(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	location := &domain.Location{Name: "Airport", Address: "123 Airport Rd", Capacity: 50}
	repo.On("CreateLocation", mock.Anything, location).Return(nil)
	cache.On("SetLocation", mock.Anything, location).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	created, err := svc.CreateLocation(context.Background(), location)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestVehicleService_CreateMaintenanceRecord_SetsVehicleStatus(t *testing.T) {
	repo := new(mockVehicleRepo)
	cache := new(mockCacheRepo)

	record := &domain.MaintenanceRecord{
		VehicleID:   "v1",
		StartDate:   "2024-01-01",
		Description: "Oil change",
		Status:      "active",
	}

	repo.On("CreateMaintenanceRecord", mock.Anything, record).Return(nil)
	repo.On("UpdateVehicleStatus", mock.Anything, "v1", "under maintenance").Return(nil)
	cache.On("GetVehicle", mock.Anything, "v1").Return((*domain.Vehicle)(nil), nil)
	repo.On("GetVehicle", mock.Anything, "v1").Return(&domain.Vehicle{ID: "v1"}, nil)
	cache.On("SetVehicle", mock.Anything, mock.AnythingOfType("*domain.Vehicle")).Return(nil)

	svc := service.NewVehicleService(repo, cache, nil)
	created, err := svc.CreateMaintenanceRecord(context.Background(), record)

	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	repo.AssertExpectations(t)
}
