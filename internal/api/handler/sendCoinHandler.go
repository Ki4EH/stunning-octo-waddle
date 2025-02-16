package handler

import (
	"errors"
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (r *CombinedRepository) SendCoinHandler(c echo.Context) error {
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
	senderUserID := claims.UserID

	var sendCoinRequest models.SendCoin
	if err := c.Bind(&sendCoinRequest); err != nil || sendCoinRequest.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"errors": "invalid request"})
	}

	receiverChan := make(chan *models.Credential, 1)
	errChan := make(chan error, 1)
	go func() {
		receiver, err := r.userRepo.GetUserCredentialByName(c.Request().Context(), sendCoinRequest.ToUser)
		if receiver.ID == uuid.Nil {
			err = errors.New("receiver not found")
		}
		if err != nil {
			errChan <- err
			return
		}
		receiverChan <- receiver
	}()

	select {
	case receiver := <-receiverChan:
		if receiver.ID == senderUserID {
			return c.JSON(http.StatusBadRequest, map[string]string{"errors": "cannot send coins to yourself"})
		}

		if err := r.coinRepo.SendCoins(c.Request().Context(), senderUserID, receiver.ID, sendCoinRequest.Amount); err != nil {
			c.Logger().Error("failed to send coins", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"errors": fmt.Sprintf("failed to send coins %v", err)})
		}

		return c.NoContent(http.StatusOK)

	case err := <-errChan:
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, map[string]string{"errors": "receiver not found"})
		}
		c.Logger().Error("failed to fetch receiver info", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"errors": fmt.Sprintf("failed to fetch receiver info %v", err)})

	case <-c.Request().Context().Done():
		c.Logger().Error("request canceled or timed out")
		return c.JSON(http.StatusInternalServerError, map[string]string{"errors": "request canceled or timed out"})
	}
}
