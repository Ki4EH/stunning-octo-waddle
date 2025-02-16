package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CoinRepository interface {
	BuyItemFromShop(ctx context.Context, userID uuid.UUID, itemName string) error
	SendCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int64) error
	GetTransactions(ctx context.Context, userID uuid.UUID, transactions *[]models.Transaction) error
}

type coinRepository struct {
	db *pgxpool.Pool
}

func NewCoinRepository(db *pgxpool.Pool) CoinRepository {
	return &coinRepository{
		db: db,
	}
}

func (r *coinRepository) BuyItemFromShop(ctx context.Context, userID uuid.UUID, itemName string) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback(ctx)
		}
	}()

	if err = r.processTransaction(ctx, tx, userID, itemName); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (r *coinRepository) processTransaction(ctx context.Context, tx pgx.Tx, userID uuid.UUID, itemName string) error {
	user, err := r.getUser(ctx, tx, userID)
	if err != nil {
		return err
	}

	shop, err := r.getShopItem(ctx, tx, itemName)
	if err != nil {
		return err
	}

	if err = r.validateBalance(user, shop.Price); err != nil {
		return err
	}

	if err = r.updateUserBalance(ctx, tx, user, shop.Price); err != nil {
		return err
	}

	if err = r.updateUserInventory(ctx, tx, userID, shop.Item); err != nil {
		return err
	}

	return nil
}

func (r *coinRepository) getUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*models.Credential, error) {
	var user models.Credential
	err := tx.QueryRow(ctx, "SELECT id, username, coin FROM credentials WHERE id = $1", userID).
		Scan(&user.ID, &user.Username, &user.Coin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *coinRepository) getShopItem(ctx context.Context, tx pgx.Tx, itemName string) (*models.Shop, error) {
	var shop models.Shop
	err := tx.QueryRow(ctx, "SELECT item, price FROM shops WHERE item = $1", itemName).
		Scan(&shop.Item, &shop.Price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("item not found")
		}
		return nil, err
	}
	return &shop, nil
}

func (r *coinRepository) validateBalance(user *models.Credential, price int64) error {
	if user.Coin < price {
		return errors.New("insufficient balance")
	}
	return nil
}

func (r *coinRepository) updateUserBalance(ctx context.Context, tx pgx.Tx, user *models.Credential, price int64) error {
	_, err := tx.Exec(ctx, "UPDATE credentials SET coin = coin - $1 WHERE id = $2", price, user.ID)
	if err != nil {
		return errors.New("failed to update user balance")
	}
	return nil
}

func (r *coinRepository) updateUserInventory(ctx context.Context, tx pgx.Tx, userID uuid.UUID, itemType string) error {
	var quantity int
	err := tx.QueryRow(ctx, "SELECT quantity FROM user_items WHERE user_id = $1 AND type = $2", userID, itemType).Scan(&quantity)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = tx.Exec(ctx, "INSERT INTO user_items (user_id, type, quantity) VALUES ($1, $2, $3)", userID, itemType, 1)
			if err != nil {
				return errors.New("failed to add item to inventory")
			}
			return nil
		}
		return errors.New("failed to check user inventory")
	}

	_, err = tx.Exec(ctx, "UPDATE user_items SET quantity = quantity + 1 WHERE user_id = $1 AND type = $2", userID, itemType)
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %v", err)
	}

	return nil
}

func (r *coinRepository) SendCoins(ctx context.Context, fromUserID, toUserID uuid.UUID, amount int64) error {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback(ctx)
		}
	}()

	var fromUser models.Credential
	err = tx.QueryRow(ctx, "SELECT id, coin FROM credentials WHERE id = $1", fromUserID).
		Scan(&fromUser.ID, &fromUser.Coin)
	if err != nil {
		tx.Rollback(ctx)
		return errors.New("sender not found")
	}
	if fromUser.Coin < amount {
		tx.Rollback(ctx)
		return errors.New("insufficient balance")
	}

	_, err = tx.Exec(ctx, "UPDATE credentials SET coin = coin - $1 WHERE id = $2", amount, fromUserID)
	if err != nil {
		tx.Rollback(ctx)
		return errors.New("failed to update sender balance")
	}

	_, err = tx.Exec(ctx, "UPDATE credentials SET coin = coin + $1 WHERE id = $2", amount, toUserID)
	if err != nil {
		tx.Rollback(ctx)
		return errors.New("failed to update receiver balance")
	}

	_, err = tx.Exec(ctx, "INSERT INTO transactions (from_user, to_user, amount) VALUES ($1, $2, $3)", fromUserID, toUserID, amount)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to record transaction: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return errors.New("failed to commit transaction")
	}
	return nil
}

func (r *coinRepository) GetTransactions(ctx context.Context, userID uuid.UUID, transactions *[]models.Transaction) error {
	query := `
		SELECT t.id, cf.username AS from_user, ct.username AS to_user, t.amount
		FROM transactions t
		JOIN credentials cf ON t.from_user = cf.id
		JOIN credentials ct ON t.to_user = ct.id
		WHERE t.from_user = $1 OR t.to_user = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		if err = rows.Scan(&t.ID, &t.FromUser, &t.ToUser, &t.Amount); err != nil {
			return err
		}
		*transactions = append(*transactions, t)
	}

	return nil
}
