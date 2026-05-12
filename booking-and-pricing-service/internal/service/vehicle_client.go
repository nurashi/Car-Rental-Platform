package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	inventorypb "github.com/nurashi/car-rental-booking/internal/clients/inventory/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// VehicleClient is the gRPC client wrapper for the Vehicle Inventory Service.
// It implements VehicleServiceClient.
type VehicleClient struct {
	client inventorypb.VehicleInventoryServiceClient
	conn   *grpc.ClientConn
}

// Ensure VehicleClient satisfies the interface.
var _ VehicleServiceClient = (*VehicleClient)(nil)

func NewVehicleClient(addr string) (*VehicleClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to vehicle service: %w", err)
	}
	log.Printf("connected to vehicle inventory service at %s", addr)
	return &VehicleClient{
		client: inventorypb.NewVehicleInventoryServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *VehicleClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// GetVehicleCategory calls GetVehicle and derives a pricing category from the model name.
// NOTE: Coordinate with the vehicle-inventory-service team to add an explicit
// "category" field to the Vehicle proto message to replace this heuristic.
func (c *VehicleClient) GetVehicleCategory(ctx context.Context, vehicleID string) (string, error) {
	resp, err := c.client.GetVehicle(ctx, &inventorypb.GetVehicleRequest{Id: vehicleID})
	if err != nil {
		return "economy", fmt.Errorf("get vehicle: %w", err)
	}

	model := strings.ToLower(resp.GetModel())
	switch {
	case containsAny(model, "suv", "4x4", "jeep", "ranger", "explorer", "patrol"):
		return "suv", nil
	case containsAny(model, "bmw", "mercedes", "audi", "porsche", "lexus", "infinity", "luxury"):
		return "luxury", nil
	case containsAny(model, "compact", "polo", "fiesta", "yaris", "corsa", "clio"):
		return "compact", nil
	default:
		return "economy", nil
	}
}

// IsVehicleAvailable returns true when the vehicle status equals "available".
func (c *VehicleClient) IsVehicleAvailable(ctx context.Context, vehicleID string) (bool, error) {
	resp, err := c.client.GetVehicle(ctx, &inventorypb.GetVehicleRequest{Id: vehicleID})
	if err != nil {
		return false, fmt.Errorf("get vehicle: %w", err)
	}
	return strings.EqualFold(resp.GetStatus(), "available"), nil
}

// CheckVehicleAvailability implements Graceful Degradation when checking the inventory.
func (c *VehicleClient) CheckVehicleAvailability(ctx context.Context, vehicleID string) (bool, float64) {
	resp, err := c.client.GetVehicle(ctx, &inventorypb.GetVehicleRequest{Id: vehicleID})
	
	if err != nil {
		// --- THIS IS THE FALLBACK ---
		log.Printf("Inventory Service is DOWN: %v. Using fallback defaults.", err)
		return true, 50.0 // Default to "Available" and "Economy Price ($50)"
	}

	// Note: Teammate's protobuf has `status` instead of `IsAvailable`
	// and doesn't have a `DailyRate` field natively. So we derive it gracefully.
	isAvailable := strings.EqualFold(resp.GetStatus(), "available")
	
	// Map to a default daily rate based on model heuristics
	dailyRate := 50.0 // default economy
	model := strings.ToLower(resp.GetModel())
	switch {
	case containsAny(model, "suv", "4x4", "jeep"):
		dailyRate = 80.0
	case containsAny(model, "bmw", "mercedes", "audi", "luxury"):
		dailyRate = 120.0
	}

	return isAvailable, dailyRate
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
