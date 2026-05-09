package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
)

type CacheRepository interface {
	SetVehicle(ctx context.Context, vehicle *domain.Vehicle) error
	GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error)
	DeleteVehicle(ctx context.Context, id string) error
	
	SetLocation(ctx context.Context, location *domain.Location) error
	GetLocation(ctx context.Context, id string) (*domain.Location, error)
	DeleteLocation(ctx context.Context, id string) error
}

type redisCacheRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCacheRepository(client *redis.Client, ttl time.Duration) CacheRepository {
	return &redisCacheRepository{client: client, ttl: ttl}
}

func (r *redisCacheRepository) SetVehicle(ctx context.Context, vehicle *domain.Vehicle) error {
	data, err := json.Marshal(vehicle)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "vehicle:"+vehicle.ID, data, r.ttl).Err()
}

func (r *redisCacheRepository) GetVehicle(ctx context.Context, id string) (*domain.Vehicle, error) {
	data, err := r.client.Get(ctx, "vehicle:"+id).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, err
	}
	var vehicle domain.Vehicle
	if err := json.Unmarshal(data, &vehicle); err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *redisCacheRepository) DeleteVehicle(ctx context.Context, id string) error {
	return r.client.Del(ctx, "vehicle:"+id).Err()
}

func (r *redisCacheRepository) SetLocation(ctx context.Context, location *domain.Location) error {
	data, err := json.Marshal(location)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "location:"+location.ID, data, r.ttl).Err()
}

func (r *redisCacheRepository) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	data, err := r.client.Get(ctx, "location:"+id).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, err
	}
	var location domain.Location
	if err := json.Unmarshal(data, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

func (r *redisCacheRepository) DeleteLocation(ctx context.Context, id string) error {
	return r.client.Del(ctx, "location:"+id).Err()
}
