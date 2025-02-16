package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/mi4r/avito-shop/internal/handlers"
	"github.com/mi4r/avito-shop/internal/storage/mocks"
	"github.com/mi4r/avito-shop/internal/storage/models"
	"github.com/mi4r/avito-shop/internal/storage/storage"
)

func TestAuthHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        models.AuthRequest
		mockSetup      func(*mocks.Storage)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			request: models.AuthRequest{
				Username: "newuser",
				Password: "password",
			},
			mockSetup: func(m *mocks.Storage) {
				m.On("GetUserByUsername", mock.Anything, "newuser").
					Return((*models.User)(nil), sql.ErrNoRows)
				m.On("CreateUser", mock.Anything, "newuser", mock.Anything).
					Return(&models.User{Username: "newuser"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful login",
			request: models.AuthRequest{
				Username: "existinguser",
				Password: "correctpassword",
			},
			mockSetup: func(m *mocks.Storage) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				m.On("GetUserByUsername", mock.Anything, "existinguser").
					Return(&models.User{
						Username:     "existinguser",
						PasswordHash: string(hash),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			request: models.AuthRequest{
				Username: "existinguser",
				Password: "wrongpassword",
			},
			mockSetup: func(m *mocks.Storage) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				m.On("GetUserByUsername", mock.Anything, "existinguser").
					Return(&models.User{
						Username:     "existinguser",
						PasswordHash: string(hash),
					}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid password",
		},
		{
			name: "create user error",
			request: models.AuthRequest{
				Username: "newuser",
				Password: "password",
			},
			mockSetup: func(m *mocks.Storage) {
				m.On("GetUserByUsername", mock.Anything, "newuser").
					Return((*models.User)(nil), sql.ErrNoRows)
				m.On("CreateUser", mock.Anything, "newuser", mock.Anything).
					Return((*models.User)(nil), errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewStorage(t)
			tt.mockSetup(mockStorage)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/auth", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler := handlers.AuthHandler(mockStorage, []byte("secret"))
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]string
				json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestInfoHandler(t *testing.T) {
	t.Run("successful info retrieval", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		expectedUser := &models.User{
			ID:    1,
			Coins: 1000,
		}

		mockStorage.On("GetUserByUsername", mock.Anything, "testuser").
			Return(expectedUser, nil)
		mockStorage.On("GetUserInventory", mock.Anything, 1).
			Return([]models.InventoryItem{}, nil)
		mockStorage.On("GetCoinHistory", mock.Anything, 1).
			Return([]models.ReceivedTransaction{}, []models.SentTransaction{}, nil)

		req := httptest.NewRequest("GET", "/info", nil)
		ctx := context.WithValue(req.Context(), "username", "testuser")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := handlers.InfoHandler(mockStorage)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.InfoResponse
		json.Unmarshal(rr.Body.Bytes(), &response)
		assert.Equal(t, expectedUser.Coins, response.Coins)
	})

	t.Run("user not found", func(t *testing.T) {
		mockStorage := mocks.NewStorage(t)
		mockStorage.On("GetUserByUsername", mock.Anything, "testuser").
			Return((*models.User)(nil), errors.New("not found"))

		req := httptest.NewRequest("GET", "/info", nil)
		ctx := context.WithValue(req.Context(), "username", "testuser")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler := handlers.InfoHandler(mockStorage)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestSendCoinHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        models.SendCoinRequest
		mockSetup      func(*mocks.Storage)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful transfer",
			request: models.SendCoinRequest{
				ToUser: "receiver",
				Amount: 100,
			},
			mockSetup: func(m *mocks.Storage) {
				m.On("SendCoins", mock.Anything, "sender", "receiver", 100).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "insufficient coins",
			request: models.SendCoinRequest{
				ToUser: "receiver",
				Amount: 1000,
			},
			mockSetup: func(m *mocks.Storage) {
				m.On("SendCoins", mock.Anything, "sender", "receiver", 1000).
					Return(storage.ErrInsufficientCoins)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "insufficient coins",
		},
		{
			name: "user not found",
			request: models.SendCoinRequest{
				ToUser: "nonexistent",
				Amount: 100,
			},
			mockSetup: func(m *mocks.Storage) {
				m.On("SendCoins", mock.Anything, "sender", "nonexistent", 100).
					Return(storage.ErrUserNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewStorage(t)
			tt.mockSetup(mockStorage)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/send", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), "username", "sender")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := handlers.SendCoinHandler(mockStorage)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]string
				json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBuyItemHandler(t *testing.T) {
	tests := []struct {
		name           string
		itemName       string
		mockSetup      func(*mocks.Storage)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful purchase",
			itemName: "t-shirt",
			mockSetup: func(m *mocks.Storage) {
				m.On("BuyItem", mock.Anything, "buyer", "t-shirt").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "item not found",
			itemName: "invalid",
			mockSetup: func(m *mocks.Storage) {
				m.On("BuyItem", mock.Anything, "buyer", "invalid").
					Return(storage.ErrItemNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "item not found",
		},
		{
			name:     "insufficient coins",
			itemName: "expensive",
			mockSetup: func(m *mocks.Storage) {
				m.On("BuyItem", mock.Anything, "buyer", "expensive").
					Return(storage.ErrInsufficientCoins)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "insufficient coins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewStorage(t)
			tt.mockSetup(mockStorage)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("item", tt.itemName)

			req := httptest.NewRequest("GET", "/buy/"+tt.itemName, nil)
			ctx := context.WithValue(req.Context(), "username", "buyer")
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := handlers.BuyItemHandler(mockStorage)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]string
				json.Unmarshal(rr.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
