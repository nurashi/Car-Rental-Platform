package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/messaging"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/repository"
)

type VehicleService interface {
	CreateVehicle(ctx context.Context, req *domain.Vehicle) (*domain.Vehicle, error)
	GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error)
	UpdateVehicle(ctx context.Context, req *domain.Vehicle) (*domain.Vehicle, error)
	DeleteVehicle(ctx context.Context, id string) error
	ListVehicles(ctx context.Context, limit, offset int) ([]*domain.Vehicle, int, error)
	UpdateVehicleStatus(ctx context.Context, id string, status string) (*domain.Vehicle, error)
	
	CreateLocation(ctx context.Context, req *domain.Location) (*domain.Location, error)
	GetLocation(ctx context.Context, id string) (*domain.Location, error)
	ListLocations(ctx context.Context, limit, offset int) ([]*domain.Location, int, error)

	CreateMaintenanceRecord(ctx context.Context, req *domain.MaintenanceRecord) (*domain.MaintenanceRecord, error)
	GetMaintenanceRecord(ctx context.Context, id string) (*domain.MaintenanceRecord, error)
	ListMaintenanceRecords(ctx context.Context, vehicleID string, limit, offset int) ([]*domain.MaintenanceRecord, int, error)
}

type vehicleService struct {
	dbRepo    repository.VehicleRepository
	cacheRepo repository.CacheRepository
	nats      *messaging.NATSClient
}

func NewVehicleService(dbRepo repository.VehicleRepository, cacheRepo repository.CacheRepository, nats *messaging.NATSClient) VehicleService {
	return &vehicleService{
		dbRepo:    dbRepo,
		cacheRepo: cacheRepo,
		nats:      nats,
	}
}

func (s *vehicleService) CreateVehicle(ctx context.Context, vehicle *domain.Vehicle) (*domain.Vehicle, error) {
	vehicle.ID = uuid.New().String()
	if err := s.dbRepo.CreateVehicle(ctx, vehicle); err != nil {
		return nil, err
	}
	s.cacheRepo.SetVehicle(ctx, vehicle)
	s.publishEvent("vehicle.created", vehicle)
	return vehicle, nil
}

func (s *vehicleService) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	if cached, _ := s.cacheRepo.GetVehicle(ctx, id); cached != nil {
		return cached, nil
	}
	
	vehicle, err := s.dbRepo.GetVehicle(ctx, id)
	if err != nil {
		return nil, err
	}
	s.cacheRepo.SetVehicle(ctx, vehicle)
	return vehicle, nil
}

func (s *vehicleService) UpdateVehicle(ctx context.Context, req *domain.Vehicle) (*domain.Vehicle, error) {
	if err := s.dbRepo.UpdateVehicle(ctx, req); err != nil {
		return nil, err
	}
	
	s.cacheRepo.SetVehicle(ctx, req)
	s.publishEvent("vehicle.updated", req)
	return req, nil
}

func (s *vehicleService) DeleteVehicle(ctx context.Context, id string) error {
	if err := s.dbRepo.DeleteVehicle(ctx, id); err != nil {
		return err
	}
	s.cacheRepo.DeleteVehicle(ctx, id)
	s.publishEvent("vehicle.deleted", map[string]string{"id": id})
	return nil
}

func (s *vehicleService) ListVehicles(ctx context.Context, limit, offset int) ([]*domain.Vehicle, int, error) {
	return s.dbRepo.ListVehicles(ctx, limit, offset)
}

func (s *vehicleService) UpdateVehicleStatus(ctx context.Context, id string, status string) (*domain.Vehicle, error) {
	if err := s.dbRepo.UpdateVehicleStatus(ctx, id, status); err != nil {
		return nil, err
	}
	
	vehicle, err := s.GetVehicle(ctx, id)
	if err != nil {
		return nil, err
	}
	
	s.cacheRepo.SetVehicle(ctx, vehicle)
	s.publishEvent("vehicle.status_changed", vehicle)
	return vehicle, nil
}

func (s *vehicleService) CreateLocation(ctx context.Context, location *domain.Location) (*domain.Location, error) {
	location.ID = uuid.New().String()
	if err := s.dbRepo.CreateLocation(ctx, location); err != nil {
		return nil, err
	}
	s.cacheRepo.SetLocation(ctx, location)
	return location, nil
}

func (s *vehicleService) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	if cached, _ := s.cacheRepo.GetLocation(ctx, id); cached != nil {
		return cached, nil
	}
	
	location, err := s.dbRepo.GetLocation(ctx, id)
	if err != nil {
		return nil, err
	}
	s.cacheRepo.SetLocation(ctx, location)
	return location, nil
}

func (s *vehicleService) ListLocations(ctx context.Context, limit, offset int) ([]*domain.Location, int, error) {
	return s.dbRepo.ListLocations(ctx, limit, offset)
}

func (s *vehicleService) CreateMaintenanceRecord(ctx context.Context, record *domain.MaintenanceRecord) (*domain.MaintenanceRecord, error) {
	record.ID = uuid.New().String()
	if err := s.dbRepo.CreateMaintenanceRecord(ctx, record); err != nil {
		return nil, err
	}
	
	// If a maintenance record is created, the vehicle status likely changes to "under maintenance"
	if record.Status == "active" {
		_, err := s.UpdateVehicleStatus(ctx, record.VehicleID, "under maintenance")
		if err != nil {
			fmt.Println("failed to update vehicle status:", err)
		}
	}
	
	s.publishEvent("maintenance.created", record)
	return record, nil
}

func (s *vehicleService) GetMaintenanceRecord(ctx context.Context, id string) (*domain.MaintenanceRecord, error) {
	return s.dbRepo.GetMaintenanceRecord(ctx, id)
}

func (s *vehicleService) ListMaintenanceRecords(ctx context.Context, vehicleID string, limit, offset int) ([]*domain.MaintenanceRecord, int, error) {
	return s.dbRepo.ListMaintenanceRecords(ctx, vehicleID, limit, offset)
}

func (s *vehicleService) publishEvent(subject string, data interface{}) {
	if s.nats == nil {
		return
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("failed to marshal event data: %v\n", err)
		return
	}
	if err := s.nats.Publish(subject, bytes); err != nil {
		fmt.Printf("failed to publish event: %v\n", err)
	}
}
