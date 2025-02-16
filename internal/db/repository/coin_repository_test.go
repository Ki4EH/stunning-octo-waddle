package repository

import (
	"context"
	"errors"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	mock_repository "github.com/Ki4EH/stunning-octo-waddle/internal/mocks/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestInterfaceBuyItemFromShop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockCoinRepository(ctrl)
	ctx := context.Background()
	userID := uuid.New()
	itemName := "cup"

	t.Run("successful purchase", func(t *testing.T) {
		mockRepo.EXPECT().
			BuyItemFromShop(ctx, userID, itemName).
			Return(nil)

		err := mockRepo.BuyItemFromShop(ctx, userID, itemName)
		assert.NoError(t, err)
	})

	t.Run("item not found", func(t *testing.T) {
		mockRepo.EXPECT().
			BuyItemFromShop(ctx, userID, "invalid_item").
			Return(errors.New("item not found"))

		err := mockRepo.BuyItemFromShop(ctx, userID, "invalid_item")
		assert.ErrorContains(t, err, "item not found")
	})
}

func TestInterfaceSendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockCoinRepository(ctrl)
	ctx := context.Background()
	fromUser := uuid.New()
	toUser := uuid.New()
	amount := int64(100)

	t.Run("successful transfer", func(t *testing.T) {
		mockRepo.EXPECT().
			SendCoins(ctx, fromUser, toUser, amount).
			Return(nil)

		err := mockRepo.SendCoins(ctx, fromUser, toUser, amount)
		assert.NoError(t, err)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		mockRepo.EXPECT().
			SendCoins(ctx, fromUser, toUser, amount).
			Return(errors.New("insufficient balance"))

		err := mockRepo.SendCoins(ctx, fromUser, toUser, amount)
		assert.ErrorContains(t, err, "insufficient balance")
	})
}

func TestInterfaceGetTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockCoinRepository(ctrl)
	ctx := context.Background()
	userID := uuid.New()
	expectedTransactions := []models.Transaction{
		{
			ID:       uuid.New(),
			FromUser: "user1",
			ToUser:   "user2",
			Amount:   50,
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.EXPECT().
			GetTransactions(ctx, userID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ uuid.UUID, txs *[]models.Transaction) error {
				*txs = append(*txs, expectedTransactions...)
				return nil
			})

		var transactions []models.Transaction
		err := mockRepo.GetTransactions(ctx, userID, &transactions)
		assert.NoError(t, err)
		assert.Equal(t, expectedTransactions, transactions)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetTransactions(ctx, userID, gomock.Any()).
			Return(errors.New("database error"))

		var transactions []models.Transaction
		err := mockRepo.GetTransactions(ctx, userID, &transactions)
		assert.ErrorContains(t, err, "database error")
	})
}

func setupCoin(t *testing.T) (repo *coinRepository, ctx context.Context) {
	ctx = context.Background()

	repo = &coinRepository{db: pool}

	return repo, ctx
}

func TestBuyItemFromShop(t *testing.T) {
	repo, ctx := setupCoin(t)

	t.Run("successful purchase", func(t *testing.T) {
		userID := uuid.New()
		itemName := "cup"
		price := int64(20)
		userName := uuid.NewString()

		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password, coin)
            VALUES ($1, $2, 'TestBuyItemFromShop', $3)
        `, userID, userName, price+50)
		require.NoError(t, err)

		err = repo.BuyItemFromShop(ctx, userID, itemName)
		require.NoError(t, err)

		var balance int64
		err = repo.db.QueryRow(ctx, "SELECT coin FROM credentials WHERE id = $1", userID).Scan(&balance)
		require.NoError(t, err)
		require.Equal(t, int64(50), balance)

		var quantity int
		err = repo.db.QueryRow(ctx, `
            SELECT quantity FROM user_items WHERE user_id = $1 AND type = $2
        `, userID, itemName).Scan(&quantity)
		require.NoError(t, err)
		require.Equal(t, 1, quantity)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		userID := uuid.New()
		itemName := "pink-hoody"
		price := int64(500)
		userName := uuid.NewString()

		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password, coin)
            VALUES ($1, $2, 'insufficient_balance', $3)
        `, userID, userName, price-50)
		require.NoError(t, err)

		err = repo.BuyItemFromShop(ctx, userID, itemName)
		require.ErrorContains(t, err, "insufficient balance")
	})
}

func TestSendCoins(t *testing.T) {
	repo, ctx := setupCoin(t)

	t.Run("successful transfer", func(t *testing.T) {
		fromUser := uuid.New()
		toUser := uuid.New()
		amount := int64(50)

		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password, coin)
            VALUES ($1, $2, 'pass', $3), ($4, $5, 'pass', 0)
        `, fromUser, fromUser.String(), amount+100, toUser, toUser.String())
		require.NoError(t, err)

		err = repo.SendCoins(ctx, fromUser, toUser, amount)
		require.NoError(t, err)

		var fromBalance, toBalance int64
		err = repo.db.QueryRow(ctx, "SELECT coin FROM credentials WHERE id = $1", fromUser).Scan(&fromBalance)
		require.NoError(t, err)
		require.Equal(t, int64(100), fromBalance)

		err = repo.db.QueryRow(ctx, "SELECT coin FROM credentials WHERE id = $1", toUser).Scan(&toBalance)
		require.NoError(t, err)
		require.Equal(t, amount, toBalance)

		var txCount int
		err = repo.db.QueryRow(ctx, `
            SELECT COUNT(*) FROM transactions 
            WHERE from_user = $1 AND to_user = $2 AND amount = $3
        `, fromUser, toUser, amount).Scan(&txCount)
		require.NoError(t, err)
		require.Equal(t, 1, txCount)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		fromUser := uuid.New()
		toUser := uuid.New()
		amount := int64(200)

		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password, coin)
            VALUES ($1, $2, 'pass', $3)
        `, fromUser, fromUser.String(), amount-50)
		require.NoError(t, err)

		err = repo.SendCoins(ctx, fromUser, toUser, amount)
		require.ErrorContains(t, err, "insufficient balance")
	})
}

func TestGetTransactions(t *testing.T) {
	repo, ctx := setupCoin(t)

	t.Run("fetch transactions", func(t *testing.T) {
		user1 := uuid.New()
		user2 := uuid.New()

		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password)
            VALUES ($1, $2,'user1'), ($3, $4, 'user2');
			`, user1, user1.String(), user2, user2.String())
		require.NoError(t, err)
		_, err = repo.db.Exec(ctx, `INSERT INTO transactions (from_user, to_user, amount)
            VALUES 
                ($1, $2, 100),
                ($2, $1, 50);
        `, user1, user2)
		require.NoError(t, err)

		var transactions []models.Transaction
		err = repo.GetTransactions(ctx, user1, &transactions)
		require.NoError(t, err)
		require.Len(t, transactions, 2)

		txMap := make(map[string]models.Transaction)
		for _, tx := range transactions {
			txMap[tx.FromUser+tx.ToUser] = tx
		}

		require.Equal(t, user2.String(), txMap[user1.String()+user2.String()].ToUser)
		require.Equal(t, user1.String(), txMap[user1.String()+user2.String()].FromUser)
		require.Equal(t, int64(100), txMap[user1.String()+user2.String()].Amount)

		require.Equal(t, user1.String(), txMap[user2.String()+user1.String()].ToUser)
		require.Equal(t, user2.String(), txMap[user2.String()+user1.String()].FromUser)
		require.Equal(t, int64(50), txMap[user2.String()+user1.String()].Amount)
	})
}
