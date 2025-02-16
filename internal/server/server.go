package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mi4r/avito-shop/internal/handlers"
	"github.com/mi4r/avito-shop/internal/middlewares"
	"github.com/mi4r/avito-shop/internal/storage/storage"
)

func NewServer(store storage.Storage) *http.Server {
	r := chi.NewRouter()

	secretKey := []byte("secret-key")
	authMiddleware := middleware.AuthMiddleware(secretKey, store)

	r.Post("/api/auth", handlers.AuthHandler(store, secretKey))

	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Get("/api/info", handlers.InfoHandler(store))
		r.Post("/api/sendCoin", handlers.SendCoinHandler(store))
		r.Get("/api/buy/{item}", handlers.BuyItemHandler(store))
	})
	return &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
}
