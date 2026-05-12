package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	upstreamURLs map[string]string
}

func New(urls map[string]string) *Handler {
	return &Handler{upstreamURLs: urls}
}

func (h *Handler) SetupRoutes(r *gin.Engine) {
	r.GET("/health", h.health)
	r.GET("/health/ready", h.readiness)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) readiness(c *gin.Context) {
	statuses := make(map[string]string)
	for name, url := range h.upstreamURLs {
		resp, err := http.Get(url + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			statuses[name] = "unhealthy"
		} else {
			statuses[name] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	c.JSON(http.StatusOK, gin.H{"ready": true, "upstreams": statuses})
}
