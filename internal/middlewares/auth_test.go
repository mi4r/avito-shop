package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/mi4r/avito-shop/internal/middlewares"
	"github.com/mi4r/avito-shop/internal/storage/mocks"
)

var testSecret = []byte("test-secret")

func TestAuthMiddleware(t *testing.T) {
	t.Run("no authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authorization header required")
		assert.False(t, handlerCalled)
	})

	t.Run("invalid token format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidToken")
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
		assert.False(t, handlerCalled)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
		assert.False(t, handlerCalled)
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(-1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(testSecret)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
		assert.False(t, handlerCalled)
	})

	t.Run("invalid claims type", func(t *testing.T) {
		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims = jwt.RegisteredClaims{}
		tokenString, _ := token.SignedString(testSecret)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token claims")
		assert.False(t, handlerCalled)
	})

	t.Run("missing username in claims", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": "admin",
		})
		tokenString, _ := token.SignedString(testSecret)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token claims")
		assert.False(t, handlerCalled)
	})

	t.Run("valid token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "testuser",
		})
		tokenString, _ := token.SignedString(testSecret)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		mockStorage := mocks.NewStorage(t)
		middleware := middleware.AuthMiddleware(testSecret, mockStorage)

		handlerCalled := false
		var contextUsername string
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			contextUsername = r.Context().Value("username").(string)
		})

		middleware(testHandler).ServeHTTP(rr, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, "testuser", contextUsername)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthMiddlewareWithDifferentSecret(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "testuser",
	})
	wrongSecret := []byte("wrong-secret")
	tokenString, _ := token.SignedString(wrongSecret)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	mockStorage := mocks.NewStorage(t)
	middleware := middleware.AuthMiddleware(testSecret, mockStorage)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	middleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token")
	assert.False(t, handlerCalled)
}

func TestAuthMiddlewareWithInvalidTokenType(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"username": "testuser",
	})
	tokenString, _ := token.SignedString(testSecret)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	mockStorage := mocks.NewStorage(t)
	middleware := middleware.AuthMiddleware(testSecret, mockStorage)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	middleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid token")
	assert.False(t, handlerCalled)
}
