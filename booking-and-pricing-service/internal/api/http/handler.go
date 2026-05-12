package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nurashi/car-rental-booking/internal/domain"
	"github.com/nurashi/car-rental-booking/internal/service"
)

type bookingService interface {
	CreateBooking(ctx context.Context, input service.CreateBookingInput) (*domain.Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error)
	ListUserBookings(ctx context.Context, userID string, page, pageSize int, status string) ([]domain.Booking, int, error)
	CancelBooking(ctx context.Context, bookingID, reason string) (bool, error)
	CalculatePrice(ctx context.Context, vehicleID string, start, end time.Time) (*service.PriceCalculation, error)
	CheckAvailability(ctx context.Context, vehicleID string, start, end time.Time) (*service.AvailabilityResult, error)
}

// Handler is the Gin HTTP handler for the REST API Gateway interface.
type Handler struct {
	bookingSvc bookingService
	jwtSecret  string
	engine     *gin.Engine
}

func NewHandler(bookingSvc bookingService, jwtSecret string) *Handler {
	engine := gin.Default()
	h := &Handler{bookingSvc: bookingSvc, jwtSecret: jwtSecret, engine: engine}
	h.setupRoutes()
	return h
}

func (h *Handler) Engine() *gin.Engine {
	return h.engine
}

func (h *Handler) setupRoutes() {
	api := h.engine.Group("/api/v1")

	// Public
	api.POST("/bookings/calculate-price", h.calculatePrice)
	api.GET("/bookings/availability", h.checkAvailability)

	// Protected
	protected := api.Group("/")
	protected.Use(h.authMiddleware())
	{
		protected.POST("/bookings", h.createBooking)
		protected.GET("/bookings/:id", h.getBooking)
		protected.GET("/users/:user_id/bookings", h.listUserBookings)
		protected.DELETE("/bookings/:id", h.cancelBooking)
	}
}

// ── Request/response types ────────────────────────────────────────────────────

type createBookingRequest struct {
	VehicleID         string `json:"vehicle_id" binding:"required"`
	StartDate         string `json:"start_date" binding:"required"`
	EndDate           string `json:"end_date" binding:"required"`
	PickupLocationID  string `json:"pickup_location_id"`
	DropoffLocationID string `json:"dropoff_location_id"`
	Notes             string `json:"notes"`
}

type calculatePriceRequest struct {
	VehicleID string `json:"vehicle_id" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type cancelBookingRequest struct {
	Reason string `json:"reason"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) createBooking(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format (use RFC3339)"})
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format (use RFC3339)"})
		return
	}

	input := service.CreateBookingInput{
		UserID:            userID.(string),
		VehicleID:         req.VehicleID,
		StartDate:         start,
		EndDate:           end,
		PickupLocationID:  req.PickupLocationID,
		DropoffLocationID: req.DropoffLocationID,
		Notes:             req.Notes,
	}

	b, err := h.bookingSvc.CreateBooking(c.Request.Context(), input)
	if err != nil {
		if err == service.ErrVehicleNotAvailable {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrInvalidDateRange {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		log.Printf("create booking error: %v", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking": b})
}

func (h *Handler) getBooking(c *gin.Context) {
	id := c.Param("id")
	b, err := h.bookingSvc.GetBooking(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": b})
}

func (h *Handler) listUserBookings(c *gin.Context) {
	userID := c.Param("user_id")
	page := 1
	pageSize := 20
	bookingStatus := c.Query("status")

	bookings, total, err := h.bookingSvc.ListUserBookings(c.Request.Context(), userID, page, pageSize, bookingStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bookings": bookings, "total": total})
}

func (h *Handler) cancelBooking(c *gin.Context) {
	id := c.Param("id")
	var req cancelBookingRequest
	_ = c.ShouldBindJSON(&req)

	refundable, err := h.bookingSvc.CancelBooking(c.Request.Context(), id, req.Reason)
	if err != nil {
		if err == service.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrCannotCancelBooking {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled", "refundable": refundable})
}

func (h *Handler) calculatePrice(c *gin.Context) {
	var req calculatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}

	calc, err := h.bookingSvc.CalculatePrice(c.Request.Context(), req.VehicleID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate price"})
		return
	}
	c.JSON(http.StatusOK, calc)
}

func (h *Handler) checkAvailability(c *gin.Context) {
	vehicleID := c.Query("vehicle_id")
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if vehicleID == "" || startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id, start_date, end_date are required"})
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date"})
		return
	}

	result, err := h.bookingSvc.CheckAvailability(c.Request.Context(), vehicleID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check availability"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Auth middleware ────────────────────────────────────────────────────────────

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
