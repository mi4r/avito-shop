package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mi4r/avito-shop/internal/storage/models"
	"github.com/mi4r/avito-shop/internal/storage/storage"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func AuthHandler(store storage.Storage, secretKey []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request")
			return
		}

		user, err := store.GetUserByUsername(r.Context(), req.Username)
		if errors.Is(err, sql.ErrNoRows) {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			user, err = store.CreateUser(r.Context(), req.Username, string(hashedPassword))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "failed to create user")
				return
			}
		} else if err != nil {
			respondWithError(w, http.StatusInternalServerError, "database error")
			return
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
				respondWithError(w, http.StatusUnauthorized, "invalid password")
				return
			}
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
		})
		tokenString, _ := token.SignedString(secretKey)

		respondWithJSON(w, http.StatusOK, models.AuthResponse{Token: tokenString})
	}
}

func InfoHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)

		user, err := store.GetUserByUsername(r.Context(), username)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "user not found")
			return
		}

		inventory, err := store.GetUserInventory(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to get inventory")
			return
		}

		received, sent, err := store.GetCoinHistory(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to get history")
			return
		}

		respondWithJSON(w, http.StatusOK, models.InfoResponse{
			Coins:     user.Coins,
			Inventory: inventory,
			CoinHistory: struct {
				Received []models.ReceivedTransaction "json:\"received\""
				Sent     []models.SentTransaction     "json:\"sent\""
			}{
				Received: received,
				Sent:     sent,
			},
		})
	}
}

func SendCoinHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.SendCoinRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid request")
			return
		}

		sender := r.Context().Value("username").(string)
		if err := store.SendCoins(r.Context(), sender, req.ToUser, req.Amount); err != nil {
			switch err {
			case storage.ErrInsufficientCoins:
				respondWithError(w, http.StatusBadRequest, "insufficient coins")
			case storage.ErrUserNotFound:
				respondWithError(w, http.StatusBadRequest, "user not found")
			default:
				respondWithError(w, http.StatusInternalServerError, "transaction failed")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func BuyItemHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		itemName := chi.URLParam(r, "item")
		username := r.Context().Value("username").(string)

		if err := store.BuyItem(r.Context(), username, itemName); err != nil {
			switch err {
			case storage.ErrItemNotFound:
				respondWithError(w, http.StatusBadRequest, "item not found")
			case storage.ErrInsufficientCoins:
				respondWithError(w, http.StatusBadRequest, "insufficient coins")
			default:
				respondWithError(w, http.StatusInternalServerError, "purchase failed")
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
