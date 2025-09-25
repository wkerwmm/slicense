package handler

import (
	"encoding/json"
	"net/http"

	"license-server/utils"
	"license-server/web/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username       string `json:"username"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		PasswordRepeat string `json:"passwordRepeat"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	err := h.authService.Register(body.Username, body.Email, body.Password, body.PasswordRepeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Kayıt başarılı"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	user, err := h.authService.Login(body.Email, body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Token oluşturulamadı", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Giriş başarılı",
		"token":   token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}
