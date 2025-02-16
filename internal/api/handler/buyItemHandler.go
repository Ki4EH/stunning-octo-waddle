package handler

import (
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

type CoinHandler struct {
	repo repository.CoinRepository
}

func NewCoinHandler(repo repository.CoinRepository) *CoinHandler {
	return &CoinHandler{
		repo: repo,
	}
}

func (r *CoinHandler) BuyItem(c echo.Context) error {
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
	item := c.Param("item")

	err := r.repo.BuyItemFromShop(c.Request().Context(), userID, item)

	if err != nil {
		c.Response().Status = http.StatusBadRequest
		c.Logger().Error("failed to buy item ", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"errors": fmt.Sprintf("failed to buy item %v", err)})
	}

	return c.NoContent(http.StatusOK)
}
