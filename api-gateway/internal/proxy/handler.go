package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	identityAddr string
	vehicleAddr  string
	bookingAddr  string
}

func NewHandler(identityAddr, vehicleAddr, bookingAddr string) *Handler {
	return &Handler{
		identityAddr: identityAddr,
		vehicleAddr:  vehicleAddr,
		bookingAddr:  bookingAddr,
	}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// Identity routes
	api.POST("/identity/register", h.proxyTo(h.identityAddr, "/api/v1/register"))
	api.POST("/identity/login", h.proxyTo(h.identityAddr, "/api/v1/login"))
	api.POST("/identity/verify-email", h.proxyTo(h.identityAddr, "/api/v1/verify-email"))
	api.POST("/identity/resend-verification", h.proxyTo(h.identityAddr, "/api/v1/resend-verification"))
	api.GET("/identity/profile/:id", h.proxyTo(h.identityAddr, "/api/v1/profile/:id"))
	api.PUT("/identity/profile/:id", h.proxyTo(h.identityAddr, "/api/v1/profile/:id"))

	// Vehicle routes
	api.GET("/vehicles", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles"))
	api.GET("/vehicles/:id", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id"))
	api.POST("/vehicles", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles"))
	api.PUT("/vehicles/:id", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id"))
	api.DELETE("/vehicles/:id", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id"))
	api.PUT("/vehicles/:id/status", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id/status"))

	api.GET("/locations", h.proxyTo(h.vehicleAddr, "/api/v1/locations"))
	api.GET("/locations/:id", h.proxyTo(h.vehicleAddr, "/api/v1/locations/:id"))
	api.POST("/locations", h.proxyTo(h.vehicleAddr, "/api/v1/locations"))

	api.GET("/vehicles/:id/maintenance", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id/maintenance"))
	api.POST("/vehicles/:id/maintenance", h.proxyTo(h.vehicleAddr, "/api/v1/vehicles/:id/maintenance"))

	// Booking routes
	api.POST("/bookings", h.proxyTo(h.bookingAddr, "/api/v1/bookings"))
	api.GET("/bookings/:id", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id"))
	api.DELETE("/bookings/:id", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id"))
	api.GET("/users/:user_id/bookings", h.proxyTo(h.bookingAddr, "/api/v1/users/:user_id/bookings"))
	api.GET("/bookings/availability", h.proxyTo(h.bookingAddr, "/api/v1/bookings/availability"))
	api.POST("/bookings/calculate-price", h.proxyTo(h.bookingAddr, "/api/v1/bookings/calculate-price"))
	api.POST("/bookings/:id/extend", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id/extend"))
	api.POST("/bookings/:id/payment", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id/payment"))
	api.GET("/bookings/:id/refunds", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id/refunds"))
	api.GET("/users/:user_id/history", h.proxyTo(h.bookingAddr, "/api/v1/users/:user_id/history"))
	api.POST("/bookings/:id/refund", h.proxyTo(h.bookingAddr, "/api/v1/bookings/:id/refund"))
	api.GET("/pricing-tiers", h.proxyTo(h.bookingAddr, "/api/v1/pricing-tiers"))
}

func (h *Handler) proxyTo(targetAddr string, targetPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := fmt.Sprintf("http://%s%s", targetAddr, resolvePath(targetPath, c))

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, url, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to create request"})
			return
		}

		req.Header = c.Request.Header.Clone()
		req.Header.Del("Content-Length")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		c.Data(resp.StatusCode, c.Writer.Header().Get("Content-Type"), body)
	}
}

func resolvePath(pattern string, c *gin.Context) string {
	path := pattern
	for _, param := range c.Params {
		path = strings.Replace(path, ":"+param.Key, param.Value, 1)
	}
	if c.Request.URL.RawQuery != "" {
		path += "?" + c.Request.URL.RawQuery
	}
	return path
}
