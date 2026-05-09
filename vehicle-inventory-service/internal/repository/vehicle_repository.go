package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
)

type VehicleRepository interface {
	CreateVehicle(ctx context.Context, vehicle *domain.Vehicle) error
	GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error)
	UpdateVehicle(ctx context.Context, vehicle *domain.Vehicle) error
	DeleteVehicle(ctx context.Context, id string) error
	ListVehicles(ctx context.Context, limit, offset int) ([]*domain.Vehicle, int, error)
	UpdateVehicleStatus(ctx context.Context, id string, status string) error
	
	CreateLocation(ctx context.Context, location *domain.Location) error
	GetLocation(ctx context.Context, id string) (*domain.Location, error)
	ListLocations(ctx context.Context, limit, offset int) ([]*domain.Location, int, error)

	CreateMaintenanceRecord(ctx context.Context, record *domain.MaintenanceRecord) error
	GetMaintenanceRecord(ctx context.Context, id string) (*domain.MaintenanceRecord, error)
	ListMaintenanceRecords(ctx context.Context, vehicleID string, limit, offset int) ([]*domain.MaintenanceRecord, int, error)
}

type postgresVehicleRepository struct {
	db *pgxpool.Pool
}

func NewPostgresVehicleRepository(db *pgxpool.Pool) VehicleRepository {
	return &postgresVehicleRepository{db: db}
}

func (r *postgresVehicleRepository) CreateVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `INSERT INTO vehicles (id, make, model, year, license_plate, status, current_location_id, mileage) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, vehicle.ID, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.LicensePlate, vehicle.Status, vehicle.CurrentLocationID, vehicle.Mileage)
	return err
}

func (r *postgresVehicleRepository) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	query := `SELECT id, make, model, year, license_plate, status, current_location_id, mileage FROM vehicles WHERE id = $1`
	var v domain.Vehicle
	err := r.db.QueryRow(ctx, query, id).Scan(&v.ID, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.Status, &v.CurrentLocationID, &v.Mileage)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *postgresVehicleRepository) UpdateVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	query := `UPDATE vehicles SET make = $1, model = $2, year = $3, license_plate = $4, status = $5, current_location_id = $6, mileage = $7 WHERE id = $8`
	_, err := r.db.Exec(ctx, query, vehicle.Make, vehicle.Model, vehicle.Year, vehicle.LicensePlate, vehicle.Status, vehicle.CurrentLocationID, vehicle.Mileage, vehicle.ID)
	return err
}

func (r *postgresVehicleRepository) DeleteVehicle(ctx context.Context, id string) error {
	query := `DELETE FROM vehicles WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresVehicleRepository) ListVehicles(ctx context.Context, limit, offset int) ([]*domain.Vehicle, int, error) {
	query := `SELECT id, make, model, year, license_plate, status, current_location_id, mileage FROM vehicles ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vehicles []*domain.Vehicle
	for rows.Next() {
		var v domain.Vehicle
		if err := rows.Scan(&v.ID, &v.Make, &v.Model, &v.Year, &v.LicensePlate, &v.Status, &v.CurrentLocationID, &v.Mileage); err != nil {
			return nil, 0, err
		}
		vehicles = append(vehicles, &v)
	}

	var count int
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&count)
	return vehicles, count, nil
}

func (r *postgresVehicleRepository) UpdateVehicleStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE vehicles SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *postgresVehicleRepository) CreateLocation(ctx context.Context, location *domain.Location) error {
	query := `INSERT INTO locations (id, name, address, capacity) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, location.ID, location.Name, location.Address, location.Capacity)
	return err
}

func (r *postgresVehicleRepository) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	query := `SELECT id, name, address, capacity FROM locations WHERE id = $1`
	var l domain.Location
	err := r.db.QueryRow(ctx, query, id).Scan(&l.ID, &l.Name, &l.Address, &l.Capacity)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *postgresVehicleRepository) ListLocations(ctx context.Context, limit, offset int) ([]*domain.Location, int, error) {
	query := `SELECT id, name, address, capacity FROM locations ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		var l domain.Location
		if err := rows.Scan(&l.ID, &l.Name, &l.Address, &l.Capacity); err != nil {
			return nil, 0, err
		}
		locations = append(locations, &l)
	}

	var count int
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM locations").Scan(&count)
	return locations, count, nil
}

func (r *postgresVehicleRepository) CreateMaintenanceRecord(ctx context.Context, record *domain.MaintenanceRecord) error {
	query := `INSERT INTO maintenance_records (id, vehicle_id, start_date, end_date, description, cost, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, record.ID, record.VehicleID, record.StartDate, record.EndDate, record.Description, record.Cost, record.Status)
	return err
}

func (r *postgresVehicleRepository) GetMaintenanceRecord(ctx context.Context, id string) (*domain.MaintenanceRecord, error) {
	query := `SELECT id, vehicle_id, start_date, end_date, description, cost, status FROM maintenance_records WHERE id = $1`
	var m domain.MaintenanceRecord
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.VehicleID, &m.StartDate, &m.EndDate, &m.Description, &m.Cost, &m.Status)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *postgresVehicleRepository) ListMaintenanceRecords(ctx context.Context, vehicleID string, limit, offset int) ([]*domain.MaintenanceRecord, int, error) {
	query := `SELECT id, vehicle_id, start_date, end_date, description, cost, status FROM maintenance_records WHERE vehicle_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*domain.MaintenanceRecord
	for rows.Next() {
		var m domain.MaintenanceRecord
		if err := rows.Scan(&m.ID, &m.VehicleID, &m.StartDate, &m.EndDate, &m.Description, &m.Cost, &m.Status); err != nil {
			return nil, 0, err
		}
		records = append(records, &m)
	}

	var count int
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM maintenance_records WHERE vehicle_id = $1", vehicleID).Scan(&count)
	return records, count, nil
}
