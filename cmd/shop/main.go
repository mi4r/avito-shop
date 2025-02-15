package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/mi4r/avito-shop/internal/handlers"
	"github.com/mi4r/avito-shop/internal/middlewares"
	"github.com/mi4r/avito-shop/internal/storage/storage"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://mi4r:1234@localhost:5432/shop_storage?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	store := storage.NewPostgresStorage(db)

	r := chi.NewRouter()

	secretKey := []byte("secret-key")
	authMiddleware := middleware.AuthMiddleware(secretKey, store)

	r.Post("/api/auth", handlers.AuthHandler(store, secretKey))

	r.With(authMiddleware).Group(func(r chi.Router) {
		r.Get("/api/info", handlers.InfoHandler(store))
		r.Post("/api/sendCoin", handlers.SendCoinHandler(store))
		r.Get("/api/buy/{item}", handlers.BuyItemHandler(store))
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
