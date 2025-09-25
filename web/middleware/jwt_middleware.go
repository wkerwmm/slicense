package middleware

import (
	"context"
	"net/http"
	"strings"

	"license-server/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Yetkisiz (eksik token)", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			http.Error(w, "Geçersiz token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(UserIDKey).(int)
	return id, ok
}
