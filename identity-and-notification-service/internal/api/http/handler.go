package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/service"
)

type authService interface {
	Register(ctx context.Context, email, password, firstName, lastName, phone string) (string, error)
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error
	Logout(ctx context.Context, token string) error
	ValidateToken(tokenString string) (string, error)
}

type userService interface {
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, input service.UpdateProfileInput) (*domain.User, error)
}

type Handler struct {
	authService authService
	userService userService
	engine      *gin.Engine
}

func NewHandler(authSvc authService, userSvc userService) *Handler {
	engine := gin.Default()
	h := &Handler{authService: authSvc, userService: userSvc, engine: engine}
	h.setupRoutes()
	return h
}

func (h *Handler) Engine() *gin.Engine {
	return h.engine
}

func (h *Handler) setupRoutes() {
	api := h.engine.Group("/api/v1")

	api.POST("/register", h.register)
	api.POST("/login", h.login)
	api.POST("/verify-email", h.verifyEmail)
	api.POST("/resend-verification", h.resendVerification)

	protected := api.Group("/")
	protected.Use(h.authMiddleware())
	{
		protected.GET("/profile/:id", h.getProfile)
		protected.PUT("/profile/:id", h.updateProfile)
	}
}

type registerRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": userID, "message": "registration successful, check your email for verification"})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUserBlocked {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) verifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		if err == service.ErrInvalidToken {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *Handler) resendVerification(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ResendVerification(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification email sent"})
}

func (h *Handler) getProfile(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if err == service.ErrProfileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) updateProfile(c *gin.Context) {
	userID := c.Param("id")
	var input service.UpdateProfileInput
	input.UserID = userID

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		userID, err := h.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
