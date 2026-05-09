package domain

import (
	"time"
)

type Vehicle struct {
	ID                string    `json:"id"`
	Make              string    `json:"make"`
	Model             string    `json:"model"`
	Year              int       `json:"year"`
	LicensePlate      string    `json:"license_plate"`
	Status            string    `json:"status"`
	CurrentLocationID string    `json:"current_location_id"`
	Mileage           int       `json:"mileage"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Location struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MaintenanceRecord struct {
	ID          string    `json:"id"`
	VehicleID   string    `json:"vehicle_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Description string    `json:"description"`
	Cost        float64   `json:"cost"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
