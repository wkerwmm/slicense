package web

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"license-server/web/handler"
	"license-server/web/service"
)

func SetupRoutes(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/ping", handler.PingHandler)

	authService := service.NewAuthService(db)
	authHandler := handler.NewAuthHandler(authService)

	r.Route("/api/auth", func(auth chi.Router) {
		auth.Post("/register", authHandler.Register)
		auth.Post("/login", authHandler.Login)
	})

	return r
}
