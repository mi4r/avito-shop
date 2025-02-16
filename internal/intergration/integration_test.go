package integration

import (
	"database/sql"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	testDB     *sql.DB
	testRouter *http.Server
)

// TODO: fix migration default path

// func TestMain(m *testing.M) {
// 	setup()
// 	code := m.Run()
// 	teardown()
// 	os.Exit(code)
// }

// func setup() {
// 	cfg := config.NewConfig()
// 	var err error
// 	testDB, err = sql.Open("postgres", "postgres://test:test@localhost:5432/test_db?sslmode=disable")
// 	if err != nil {
// 		panic(err)
// 	}
// 	store := storage.NewPostgresStorage(testDB)
// 	store.Migrate(cfg.GetIntegDSN())
// 	testRouter = server.NewServer(store)
// }

// func teardown() {
// 	testDB.Exec("DROP TABLE IF EXISTS users, merch_items, user_inventory, coin_transactions CASCADE")
// 	testDB.Close()
// }

// func createTestUser(t *testing.T, username, password string) string {
// 	body, _ := json.Marshal(models.AuthRequest{
// 		Username: username,
// 		Password: password,
// 	})

// 	req := httptest.NewRequest("POST", "/api/auth", bytes.NewReader(body))
// 	rr := httptest.NewRecorder()
// 	testRouter.Handler.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusOK, rr.Code)

// 	var authResp models.AuthResponse
// 	json.Unmarshal(rr.Body.Bytes(), &authResp)
// 	return authResp.Token
// }

// func TestPurchaseFlow(t *testing.T) {
// 	token := createTestUser(t, "test_buyer", "password")

// 	t.Run("successful purchase", func(t *testing.T) {
// 		req := httptest.NewRequest("GET", "/api/buy/t-shirt", nil)
// 		req.Header.Set("Authorization", "Bearer "+token)
// 		rr := httptest.NewRecorder()
// 		testRouter.Handler.ServeHTTP(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)

// 		req = httptest.NewRequest("GET", "/api/info", nil)
// 		req.Header.Set("Authorization", "Bearer "+token)
// 		rr = httptest.NewRecorder()
// 		testRouter.Handler.ServeHTTP(rr, req)

// 		var info models.InfoResponse
// 		json.Unmarshal(rr.Body.Bytes(), &info)

// 		assert.Equal(t, 1000-80, info.Coins)
// 		assert.Contains(t, info.Inventory, models.InventoryItem{
// 			Type:     "t-shirt",
// 			Quantity: 1,
// 		})
// 	})
// }

// func TestCoinTransferFlow(t *testing.T) {
// 	senderToken := createTestUser(t, "sender", "password")
// 	receiverToken := createTestUser(t, "receiver", "password")

// 	t.Run("successful coin transfer", func(t *testing.T) {
// 		body, _ := json.Marshal(models.SendCoinRequest{
// 			ToUser: "receiver",
// 			Amount: 200,
// 		})

// 		req := httptest.NewRequest("POST", "/api/sendCoin", bytes.NewReader(body))
// 		req.Header.Set("Authorization", "Bearer "+senderToken)
// 		rr := httptest.NewRecorder()
// 		testRouter.Handler.ServeHTTP(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)

// 		req = httptest.NewRequest("GET", "/api/info", nil)
// 		req.Header.Set("Authorization", "Bearer "+senderToken)
// 		rr = httptest.NewRecorder()
// 		testRouter.Handler.ServeHTTP(rr, req)

// 		var senderInfo models.InfoResponse
// 		json.Unmarshal(rr.Body.Bytes(), &senderInfo)
// 		assert.Equal(t, 1000-200, senderInfo.Coins)

// 		req = httptest.NewRequest("GET", "/api/info", nil)
// 		req.Header.Set("Authorization", "Bearer "+receiverToken)
// 		rr = httptest.NewRecorder()
// 		testRouter.Handler.ServeHTTP(rr, req)

// 		var receiverInfo models.InfoResponse
// 		json.Unmarshal(rr.Body.Bytes(), &receiverInfo)
// 		assert.Equal(t, 1000+200, receiverInfo.Coins)

// 		assert.Contains(t, senderInfo.CoinHistory.Sent, models.SentTransaction{
// 			ToUser: "receiver",
// 			Amount: 200,
// 		})

// 		assert.Contains(t, receiverInfo.CoinHistory.Received, models.ReceivedTransaction{
// 			FromUser: "sender",
// 			Amount:   200,
// 		})
// 	})
// }
