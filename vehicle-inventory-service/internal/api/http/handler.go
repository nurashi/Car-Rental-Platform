package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nurashi/Car-Rental-Platform/vehicle-inventory/internal/domain"
)

type vehicleService interface {
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

type Handler struct {
	vehicleSvc vehicleService
	jwtSecret  string
	engine     *gin.Engine
}

func NewHandler(vehicleSvc vehicleService, jwtSecret string) *Handler {
	engine := gin.Default()
	h := &Handler{vehicleSvc: vehicleSvc, jwtSecret: jwtSecret, engine: engine}
	h.setupRoutes()
	return h
}

func (h *Handler) Engine() *gin.Engine {
	return h.engine
}

func (h *Handler) setupRoutes() {
	api := h.engine.Group("/api/v1")

	api.GET("/vehicles", h.listVehicles)
	api.GET("/vehicles/:id", h.getVehicle)
	api.GET("/locations", h.listLocations)
	api.GET("/locations/:id", h.getLocation)
	api.GET("/vehicles/:id/maintenance", h.listMaintenanceRecords)

	protected := api.Group("/")
	protected.Use(h.authMiddleware())
	{
		protected.POST("/vehicles", h.createVehicle)
		protected.PUT("/vehicles/:id", h.updateVehicle)
		protected.DELETE("/vehicles/:id", h.deleteVehicle)
		protected.PUT("/vehicles/:id/status", h.updateVehicleStatus)
		protected.POST("/locations", h.createLocation)
		protected.POST("/vehicles/:id/maintenance", h.createMaintenanceRecord)
	}
}

type createVehicleRequest struct {
	Make         string `json:"make" binding:"required"`
	Model        string `json:"model" binding:"required"`
	Year         int    `json:"year" binding:"required"`
	LicensePlate string `json:"license_plate" binding:"required"`
	Status       string `json:"status"`
	LocationID   string `json:"location_id"`
	Mileage      int    `json:"mileage"`
}

type updateVehicleRequest struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
	Status       string `json:"status"`
	LocationID   string `json:"location_id"`
	Mileage      int    `json:"mileage"`
}

type createLocationRequest struct {
	Name     string `json:"name" binding:"required"`
	Address  string `json:"address" binding:"required"`
	Capacity int    `json:"capacity" binding:"required"`
}

type createMaintenanceRequest struct {
	VehicleID   string  `json:"vehicle_id" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     string  `json:"end_date"`
	Description string  `json:"description" binding:"required"`
	Cost        float64 `json:"cost"`
	Status      string  `json:"status"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) createVehicle(c *gin.Context) {
	var req createVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v := &domain.Vehicle{
		Make:              req.Make,
		Model:             req.Model,
		Year:              req.Year,
		LicensePlate:      req.LicensePlate,
		Status:            req.Status,
		CurrentLocationID: req.LocationID,
		Mileage:           req.Mileage,
	}

	created, err := h.vehicleSvc.CreateVehicle(c.Request.Context(), v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create vehicle"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"vehicle": created})
}

func (h *Handler) getVehicle(c *gin.Context) {
	id := c.Param("id")
	v, err := h.vehicleSvc.GetVehicle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vehicle": v})
}

func (h *Handler) updateVehicle(c *gin.Context) {
	id := c.Param("id")
	var req updateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.vehicleSvc.GetVehicle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}

	if req.Make != "" {
		existing.Make = req.Make
	}
	if req.Model != "" {
		existing.Model = req.Model
	}
	if req.Year > 0 {
		existing.Year = req.Year
	}
	if req.LicensePlate != "" {
		existing.LicensePlate = req.LicensePlate
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.LocationID != "" {
		existing.CurrentLocationID = req.LocationID
	}
	if req.Mileage >= 0 {
		existing.Mileage = req.Mileage
	}

	updated, err := h.vehicleSvc.UpdateVehicle(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vehicle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vehicle": updated})
}

func (h *Handler) deleteVehicle(c *gin.Context) {
	id := c.Param("id")
	if err := h.vehicleSvc.DeleteVehicle(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vehicle deleted"})
}

func (h *Handler) listVehicles(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	vehicles, total, err := h.vehicleSvc.ListVehicles(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list vehicles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vehicles": vehicles, "total": total})
}

func (h *Handler) updateVehicleStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.vehicleSvc.UpdateVehicleStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vehicle": updated})
}

func (h *Handler) createLocation(c *gin.Context) {
	var req createLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	l := &domain.Location{
		Name:     req.Name,
		Address:  req.Address,
		Capacity: req.Capacity,
	}

	created, err := h.vehicleSvc.CreateLocation(c.Request.Context(), l)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create location"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"location": created})
}

func (h *Handler) getLocation(c *gin.Context) {
	id := c.Param("id")
	l, err := h.vehicleSvc.GetLocation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"location": l})
}

func (h *Handler) listLocations(c *gin.Context) {
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	locations, total, err := h.vehicleSvc.ListLocations(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list locations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations, "total": total})
}

func (h *Handler) createMaintenanceRecord(c *gin.Context) {
	var req createMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &domain.MaintenanceRecord{
		VehicleID:   req.VehicleID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: req.Description,
		Cost:        req.Cost,
		Status:      req.Status,
	}

	created, err := h.vehicleSvc.CreateMaintenanceRecord(c.Request.Context(), m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create maintenance record"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"maintenance": created})
}

func (h *Handler) getMaintenanceRecord(c *gin.Context) {
	id := c.Param("id")
	m, err := h.vehicleSvc.GetMaintenanceRecord(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "maintenance record not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"maintenance": m})
}

func (h *Handler) listMaintenanceRecords(c *gin.Context) {
	vehicleID := c.Param("id")
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	records, total, err := h.vehicleSvc.ListMaintenanceRecords(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list maintenance records"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records, "total": total})
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
