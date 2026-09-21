package handler

import (
	"net/http"

	"tokped-backend/internal/apperror"
	"tokped-backend/internal/model"
	"tokped-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Registrasi User Baru
// @Description Membuat akun shopper baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Data Registrasi"
// @Success 201 {object} model.AuthResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registrasi berhasil",
		"data":    resp,
	})
}

// Login godoc
// @Summary Login Pengguna
// @Description Autentikasi user/admin dan dapatkan JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Kredensial Login"
// @Success 200 {object} model.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil",
		"data":    resp,
	})
}

// GetProfile godoc
// @Summary Profil Pengguna
// @Description Mendapatkan informasi akun user yang sedang login
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.UserInfo
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terotentikasi"})
		return
	}

	profile, err := h.authService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profile,
	})
}
