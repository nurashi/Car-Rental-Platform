package api

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/service"
	pb "github.com/nurashi/Car-Rental-Platform/vehicle-inventory/proto"
)

type VehicleInventoryHandler struct {
	pb.UnimplementedVehicleInventoryServiceServer
	service service.VehicleService
}

func NewVehicleInventoryHandler(service service.VehicleService) *VehicleInventoryHandler {
	return &VehicleInventoryHandler{service: service}
}

func (h *VehicleInventoryHandler) CreateVehicle(ctx context.Context, req *pb.CreateVehicleRequest) (*pb.Vehicle, error) {
	v := &domain.Vehicle{
		Make:              req.Make,
		Model:             req.Model,
		Year:              int(req.Year),
		LicensePlate:      req.LicensePlate,
		Status:            req.Status,
		CurrentLocationID: req.CurrentLocationId,
		Mileage:           int(req.Mileage),
	}
	
	created, err := h.service.CreateVehicle(ctx, v)
	if err != nil {
		return nil, err
	}
	
	return mapVehicleToProto(created), nil
}

func (h *VehicleInventoryHandler) GetVehicle(ctx context.Context, req *pb.GetVehicleRequest) (*pb.Vehicle, error) {
	v, err := h.service.GetVehicle(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return mapVehicleToProto(v), nil
}

func (h *VehicleInventoryHandler) UpdateVehicle(ctx context.Context, req *pb.UpdateVehicleRequest) (*pb.Vehicle, error) {
	v := &domain.Vehicle{
		ID:                req.Id,
		Make:              req.Make,
		Model:             req.Model,
		Year:              int(req.Year),
		LicensePlate:      req.LicensePlate,
		Status:            req.Status,
		CurrentLocationID: req.CurrentLocationId,
		Mileage:           int(req.Mileage),
	}
	
	updated, err := h.service.UpdateVehicle(ctx, v)
	if err != nil {
		return nil, err
	}
	return mapVehicleToProto(updated), nil
}

func (h *VehicleInventoryHandler) DeleteVehicle(ctx context.Context, req *pb.DeleteVehicleRequest) (*emptypb.Empty, error) {
	err := h.service.DeleteVehicle(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *VehicleInventoryHandler) ListVehicles(ctx context.Context, req *pb.ListVehiclesRequest) (*pb.ListVehiclesResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}
	offset := int((req.Page - 1) * int32(limit))
	if offset < 0 {
		offset = 0
	}
	
	vehicles, total, err := h.service.ListVehicles(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	
	var pbVehicles []*pb.Vehicle
	for _, v := range vehicles {
		pbVehicles = append(pbVehicles, mapVehicleToProto(v))
	}
	
	return &pb.ListVehiclesResponse{
		Vehicles: pbVehicles,
		Total:    int32(total),
	}, nil
}

func (h *VehicleInventoryHandler) UpdateVehicleStatus(ctx context.Context, req *pb.UpdateVehicleStatusRequest) (*pb.Vehicle, error) {
	updated, err := h.service.UpdateVehicleStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return mapVehicleToProto(updated), nil
}

func (h *VehicleInventoryHandler) CreateLocation(ctx context.Context, req *pb.CreateLocationRequest) (*pb.Location, error) {
	l := &domain.Location{
		Name:     req.Name,
		Address:  req.Address,
		Capacity: int(req.Capacity),
	}
	created, err := h.service.CreateLocation(ctx, l)
	if err != nil {
		return nil, err
	}
	return mapLocationToProto(created), nil
}

func (h *VehicleInventoryHandler) GetLocation(ctx context.Context, req *pb.GetLocationRequest) (*pb.Location, error) {
	l, err := h.service.GetLocation(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return mapLocationToProto(l), nil
}

func (h *VehicleInventoryHandler) ListLocations(ctx context.Context, req *pb.ListLocationsRequest) (*pb.ListLocationsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}
	offset := int((req.Page - 1) * int32(limit))
	if offset < 0 {
		offset = 0
	}
	
	locations, total, err := h.service.ListLocations(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	
	var pbLocations []*pb.Location
	for _, l := range locations {
		pbLocations = append(pbLocations, mapLocationToProto(l))
	}
	
	return &pb.ListLocationsResponse{
		Locations: pbLocations,
		Total:     int32(total),
	}, nil
}

func (h *VehicleInventoryHandler) CreateMaintenanceRecord(ctx context.Context, req *pb.CreateMaintenanceRecordRequest) (*pb.MaintenanceRecord, error) {
	m := &domain.MaintenanceRecord{
		VehicleID:   req.VehicleId,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: req.Description,
		Cost:        req.Cost,
		Status:      req.Status,
	}
	created, err := h.service.CreateMaintenanceRecord(ctx, m)
	if err != nil {
		return nil, err
	}
	return mapMaintenanceRecordToProto(created), nil
}

func (h *VehicleInventoryHandler) GetMaintenanceRecord(ctx context.Context, req *pb.GetMaintenanceRecordRequest) (*pb.MaintenanceRecord, error) {
	m, err := h.service.GetMaintenanceRecord(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return mapMaintenanceRecordToProto(m), nil
}

func (h *VehicleInventoryHandler) ListMaintenanceRecords(ctx context.Context, req *pb.ListMaintenanceRecordsRequest) (*pb.ListMaintenanceRecordsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}
	offset := int((req.Page - 1) * int32(limit))
	if offset < 0 {
		offset = 0
	}
	
	records, total, err := h.service.ListMaintenanceRecords(ctx, req.VehicleId, limit, offset)
	if err != nil {
		return nil, err
	}
	
	var pbRecords []*pb.MaintenanceRecord
	for _, r := range records {
		pbRecords = append(pbRecords, mapMaintenanceRecordToProto(r))
	}
	
	return &pb.ListMaintenanceRecordsResponse{
		Records: pbRecords,
		Total:   int32(total),
	}, nil
}

// Helpers
func mapVehicleToProto(v *domain.Vehicle) *pb.Vehicle {
	return &pb.Vehicle{
		Id:                v.ID,
		Make:              v.Make,
		Model:             v.Model,
		Year:              int32(v.Year),
		LicensePlate:      v.LicensePlate,
		Status:            v.Status,
		CurrentLocationId: v.CurrentLocationID,
		Mileage:           int32(v.Mileage),
	}
}

func mapLocationToProto(l *domain.Location) *pb.Location {
	return &pb.Location{
		Id:       l.ID,
		Name:     l.Name,
		Address:  l.Address,
		Capacity: int32(l.Capacity),
	}
}

func mapMaintenanceRecordToProto(m *domain.MaintenanceRecord) *pb.MaintenanceRecord {
	return &pb.MaintenanceRecord{
		Id:          m.ID,
		VehicleId:   m.VehicleID,
		StartDate:   m.StartDate,
		EndDate:     m.EndDate,
		Description: m.Description,
		Cost:        m.Cost,
		Status:      m.Status,
	}
}
