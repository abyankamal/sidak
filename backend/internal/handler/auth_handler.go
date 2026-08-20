package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Payload login tidak valid")
		return
	}

	if req.NIK == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "NIK dan kata sandi wajib diisi")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(w, http.StatusUnauthorized, "NIK atau kata sandi tidak sesuai")
			return
		}
		Error(w, http.StatusInternalServerError, "Gagal memproses login pengguna")
		return
	}

	// Set HTTP-Only Cookie for Web Admin
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    resp.Token,
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	Message(w, http.StatusOK, "Logout berhasil")
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		Error(w, http.StatusUnauthorized, "Sesi tidak sah")
		return
	}

	userResp := domain.UserResponse{
		ID:   claims.UserID,
		NIK:  claims.NIK,
		Nama: claims.Nama,
		Role: claims.Role,
	}

	JSON(w, http.StatusOK, userResp)
}
