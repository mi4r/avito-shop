package main

import (
	"database/sql"
	"log"

	"github.com/mi4r/avito-shop/internal/config"
	"github.com/mi4r/avito-shop/internal/server"
	"github.com/mi4r/avito-shop/internal/storage/storage"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.NewConfig()
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	store := storage.NewPostgresStorage(db)
	store.Migrate(cfg.GetDSN())

	srv := server.NewServer(store)

	log.Println("Server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
