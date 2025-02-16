package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/mi4r/avito-shop/internal/storage/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientCoins = errors.New("insufficient coins")
	ErrItemNotFound      = errors.New("item not found")
)

type Storage interface {
	Migrate(dsn string)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error)
	GetUserInventory(ctx context.Context, userID int) ([]models.InventoryItem, error)
	GetCoinHistory(ctx context.Context, userID int) ([]models.ReceivedTransaction, []models.SentTransaction, error)
	SendCoins(ctx context.Context, senderUsername, receiverUsername string, amount int) error
	BuyItem(ctx context.Context, username, itemName string) error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (d *PostgresStorage) Migrate(dsn string) {
	// Try auto-migration
	if err := d.autoDefaultMigrate(dsn); err != nil {
		log.Fatal(err)
	}
}

func (d *PostgresStorage) autoDefaultMigrate(dsn string) error {
	mpath, err := filepath.Abs(
		filepath.Join("internal", "storage", "migrations"))
	if err != nil {
		return err
	}

	migr, err := migrate.New(
		fmt.Sprintf("file://%s", mpath),
		dsn,
	)
	if err != nil {
		return err
	}
	return migr.Up()
}

func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx, "SELECT id, username, password_hash, coins FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash) 
        VALUES ($1, $2) 
        RETURNING id, username, coins`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.Coins)

	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return nil, fmt.Errorf("username already exists")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) GetUserInventory(ctx context.Context, userID int) ([]models.InventoryItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT m.name, COALESCE(ui.quantity, 0) 
        FROM merch_items m
        LEFT JOIN user_inventory ui ON m.id = ui.item_id AND ui.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.InventoryItem
	for rows.Next() {
		var item models.InventoryItem
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, err
		}
		if item.Quantity > 0 {
			inventory = append(inventory, item)
		}
	}
	return inventory, nil
}

func (s *PostgresStorage) GetCoinHistory(ctx context.Context, userID int) ([]models.ReceivedTransaction, []models.SentTransaction, error) {
	// Получение полученных монет
	receivedRows, err := s.db.QueryContext(ctx,
		`SELECT u.username, ct.amount 
        FROM coin_transactions ct
        JOIN users u ON ct.sender_id = u.id
        WHERE ct.receiver_id = $1`,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer receivedRows.Close()

	var received []models.ReceivedTransaction
	for receivedRows.Next() {
		var t models.ReceivedTransaction
		if err := receivedRows.Scan(&t.FromUser, &t.Amount); err != nil {
			return nil, nil, err
		}
		received = append(received, t)
	}

	// Получение отправленных монет
	sentRows, err := s.db.QueryContext(ctx,
		`SELECT u.username, ct.amount 
        FROM coin_transactions ct
        JOIN users u ON ct.receiver_id = u.id
        WHERE ct.sender_id = $1`,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer sentRows.Close()

	var sent []models.SentTransaction
	for sentRows.Next() {
		var t models.SentTransaction
		if err := sentRows.Scan(&t.ToUser, &t.Amount); err != nil {
			return nil, nil, err
		}
		sent = append(sent, t)
	}

	return received, sent, nil
}

func (s *PostgresStorage) SendCoins(ctx context.Context, senderUsername, receiverUsername string, amount int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем ID отправителя и проверяем баланс
	var senderID, senderCoins int
	err = tx.QueryRowContext(ctx,
		"SELECT id, coins FROM users WHERE username = $1 FOR UPDATE",
		senderUsername,
	).Scan(&senderID, &senderCoins)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	if senderCoins < amount {
		return ErrInsufficientCoins
	}

	// Получаем ID получателя
	var receiverID int
	err = tx.QueryRowContext(ctx,
		"SELECT id FROM users WHERE username = $1 FOR UPDATE",
		receiverUsername,
	).Scan(&receiverID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	// Обновляем балансы
	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins - $1 WHERE id = $2",
		amount, senderID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins + $1 WHERE id = $2",
		amount, receiverID,
	)
	if err != nil {
		return err
	}

	// Записываем транзакцию
	_, err = tx.ExecContext(ctx,
		`INSERT INTO coin_transactions (sender_id, receiver_id, amount) 
        VALUES ($1, $2, $3)`,
		senderID, receiverID, amount,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStorage) BuyItem(ctx context.Context, username, itemName string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем информацию о товаре
	var itemID, price int
	err = tx.QueryRowContext(ctx,
		"SELECT id, price FROM merch_items WHERE name = $1",
		itemName,
	).Scan(&itemID, &price)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrItemNotFound
		}
		return err
	}

	// Получаем данные пользователя
	var userID, userCoins int
	err = tx.QueryRowContext(ctx,
		"SELECT id, coins FROM users WHERE username = $1 FOR UPDATE",
		username,
	).Scan(&userID, &userCoins)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	if userCoins < price {
		return ErrInsufficientCoins
	}

	// Обновляем баланс
	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins - $1 WHERE id = $2",
		price, userID,
	)
	if err != nil {
		return err
	}

	// Обновляем инвентарь
	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_inventory (user_id, item_id, quantity) 
        VALUES ($1, $2, 1)
        ON CONFLICT (user_id, item_id) 
        DO UPDATE SET quantity = user_inventory.quantity + 1`,
		userID, itemID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
