package handler

import (
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
	"time"
)

type CombinedRepository struct {
	userRepo repository.UserRepository
	coinRepo repository.CoinRepository
}

func NewCombinedRepository(userRepo repository.UserRepository, coinRepo repository.CoinRepository) *CombinedRepository {
	return &CombinedRepository{
		userRepo: userRepo,
		coinRepo: coinRepo,
	}
}

func (r *CombinedRepository) GetInfo(c echo.Context) error {

	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		c.Response().Status = http.StatusInternalServerError
		return c.JSON(http.StatusInternalServerError, map[string]string{"errors": "failed to get jwt token"})
	}
	claims, ok := user.Claims.(*utils.Claims)
	if !ok {
		c.Response().Status = http.StatusInternalServerError
		return c.JSON(http.StatusInternalServerError, map[string]string{"errors": "failed to get jwt token"})
	}
	userID := claims.UserID

	var wg sync.WaitGroup
	var userCredential *models.Credential
	userItems := make([]models.UserItem, 0)
	allTx := make([]models.Transaction, 0)
	var userErr, itemsErr, receivedErr, sentErr error

	wg.Add(3)
	go func() {
		defer wg.Done()

		start := time.Now()
		userCredential, userErr = r.userRepo.GetUserByID(c.Request().Context(), userID)

		elapsed := time.Since(start)
		if elapsed > 50*time.Millisecond {
			c.Logger().Error("Slow SQL", fmt.Sprintf("GetUserByID DB REQUEST took %s\n", elapsed))
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()

		itemsErr = r.userRepo.GetUserItems(c.Request().Context(), userID, &userItems)

		elapsed := time.Since(start)
		if elapsed > 50*time.Millisecond {
			c.Logger().Error("Slow SQL", fmt.Sprintf("GetUserItems DB REQUEST took %s\n", elapsed))
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()

		receivedErr = r.coinRepo.GetTransactions(c.Request().Context(), userID, &allTx)

		elapsed := time.Since(start)
		if elapsed > 50*time.Millisecond {
			c.Logger().Error("Slow SQL ", fmt.Sprintf("GetTransactions DB REQUEST took %s\n", elapsed))
		}

	}()
	wg.Wait()

	if userErr != nil || itemsErr != nil || receivedErr != nil || sentErr != nil {
		c.Logger().Error("failed to fetch data", userErr, itemsErr, receivedErr, sentErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"errors": "failed to fetch data"})
	}

	tx := models.CoinHistory{
		Received: make([]models.ReceivedTransaction, 0, len(allTx)),
		Sent:     make([]models.SentTransaction, 0, len(allTx)),
	}

	for _, t := range allTx {
		if t.ToUser == userCredential.Username {
			tx.Received = append(tx.Received, models.ReceivedTransaction{FromUser: t.FromUser, Amount: t.Amount})
		} else {
			tx.Sent = append(tx.Sent, models.SentTransaction{ToUser: t.ToUser, Amount: t.Amount})
		}
	}

	response := models.User{
		Coin:      userCredential.Coin,
		Inventory: userItems,
		CoinHistory: models.CoinHistory{
			Received: tx.Received,
			Sent:     tx.Sent,
		},
	}

	return c.JSON(http.StatusOK, response)
}
