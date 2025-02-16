package e2e_test

import (
	"encoding/json"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetInfoHandler(t *testing.T) {
	t.Run("successful user info retrieval", func(t *testing.T) {
		ctx, handler := setupCombinedTest(t)

		userID, token := createTestUser(t, 1000)

		_, err := testDB.Exec(ctx, `
            INSERT INTO user_items (user_id, type, quantity)
            VALUES ($1, 'cup', 2), ($1, 'pen', 1)
        `, userID)
		require.NoError(t, err)

		otherUserID := uuid.New()
		_, err = testDB.Exec(ctx, `
            INSERT INTO credentials (id, username, password)
            VALUES ($1, 'otheruser', 'pass');
		`, otherUserID)
		require.NoError(t, err)
		_, err = testDB.Exec(ctx, `
            INSERT INTO transactions (from_user, to_user, amount)
            VALUES 
                ($2, $1, 500),
                ($1, $2, 200)
        `, otherUserID, userID)
		require.NoError(t, err)

		e := echo.New()
		e.Use(echojwt.WithConfig(utils.JwtConfig))

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", token)
		err = handler.GetInfo(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var response models.User
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))

		require.Equal(t, int64(1000), response.Coin)
		require.Len(t, response.Inventory, 2)
		require.Len(t, response.CoinHistory.Received, 1)
		require.Len(t, response.CoinHistory.Sent, 1)

		inventoryMap := make(map[string]int64)
		for _, item := range response.Inventory {
			inventoryMap[item.Type] = item.Quantity
		}
		require.Equal(t, int64(2), inventoryMap["cup"])
		require.Equal(t, int64(1), inventoryMap["pen"])

		require.Equal(t, "otheruser", response.CoinHistory.Received[0].FromUser)
		require.Equal(t, int64(200), response.CoinHistory.Received[0].Amount)
		require.Equal(t, "otheruser", response.CoinHistory.Sent[0].ToUser)
		require.Equal(t, int64(500), response.CoinHistory.Sent[0].Amount)
	})

	t.Run("invalid JWT token", func(t *testing.T) {
		_, handler := setupCombinedTest(t)

		e := echo.New()
		e.Use(echojwt.WithConfig(utils.JwtConfig))

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handler.GetInfo(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to get jwt token")
	})

	t.Run("missing user data", func(t *testing.T) {
		_, handler := setupCombinedTest(t)

		userID := uuid.New()
		token, err := utils.GenerateToken(userID)
		require.NoError(t, err)

		e := echo.New()
		e.Use(echojwt.WithConfig(utils.JwtConfig))

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", token)
		err = handler.GetInfo(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to get jwt token")
	})

	t.Run("empty transaction history", func(t *testing.T) {
		_, handler := setupCombinedTest(t)

		_, token := createTestUser(t, 500)

		e := echo.New()
		e.Use(echojwt.WithConfig(utils.JwtConfig))

		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", token)
		err := handler.GetInfo(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var response models.User
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))

		require.Equal(t, int64(500), response.Coin)
		require.Empty(t, response.Inventory)
		require.Empty(t, response.CoinHistory.Received)
		require.Empty(t, response.CoinHistory.Sent)
	})
}
