package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Ki4EH/stunning-octo-waddle/internal/api/handler"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupCombinedTest(t *testing.T) (context.Context, *handler.CombinedRepository) {
	ctx := context.Background()

	_, err := testDB.Exec(ctx, `
        TRUNCATE credentials, transactions CASCADE;
    `)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(testDB)
	coinRepo := repository.NewCoinRepository(testDB)
	combinedRepo := handler.NewCombinedRepository(userRepo, coinRepo)

	return ctx, combinedRepo
}

func TestSendCoinHandler(t *testing.T) {
	e := echo.New()
	e.Use(echojwt.WithConfig(utils.JwtConfig))

	t.Run("successful coin transfer", func(t *testing.T) {
		ctx, handlerCombined := setupCombinedTest(t)

		senderID, senderToken := createTestUser(t, 1000)
		receiverID, _ := createTestUser(t, 0)

		requestBody := models.SendCoin{
			ToUser: receiverID.String(),
			Amount: 500,
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", senderToken)
		err := handlerCombined.SendCoinHandler(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var senderBalance, receiverBalance int64
		err = testDB.QueryRow(ctx, "SELECT coin FROM credentials WHERE id = $1", senderID).Scan(&senderBalance)
		require.NoError(t, err)
		require.Equal(t, int64(500), senderBalance)

		err = testDB.QueryRow(ctx, "SELECT coin FROM credentials WHERE id = $1", receiverID).Scan(&receiverBalance)
		require.NoError(t, err)
		require.Equal(t, int64(500), receiverBalance)

		var txCount int
		err = testDB.QueryRow(ctx, `
            SELECT COUNT(*) FROM transactions 
            WHERE from_user = $1 AND to_user = $2 AND amount = $3
        `, senderID, receiverID, 500).Scan(&txCount)
		require.NoError(t, err)
		require.Equal(t, 1, txCount)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		_, handlerCombined := setupCombinedTest(t)

		_, senderToken := createTestUser(t, 300)
		receiverID, _ := createTestUser(t, 1000)
		createTestUser(t, 0)

		requestBody := models.SendCoin{
			ToUser: receiverID.String(),
			Amount: 500,
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", senderToken)
		err := handlerCombined.SendCoinHandler(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "insufficient balance")
	})

	t.Run("send to self", func(t *testing.T) {
		_, handlerCombined := setupCombinedTest(t)

		senderID, senderToken := createTestUser(t, 1000)
		requestBody := models.SendCoin{
			ToUser: senderID.String(),
			Amount: 500,
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", senderToken)
		err := handlerCombined.SendCoinHandler(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "cannot send coins to yourself")
	})

	t.Run("receiver not found", func(t *testing.T) {
		_, handlerCombined := setupCombinedTest(t)

		_, senderToken := createTestUser(t, 1000)
		requestBody := models.SendCoin{
			ToUser: "nonexistent",
			Amount: 500,
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", senderToken)
		err := handlerCombined.SendCoinHandler(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "receiver not found")
	})

	t.Run("invalid request", func(t *testing.T) {
		_, handlerCombined := setupCombinedTest(t)

		_, senderToken := createTestUser(t, 1000)

		req := httptest.NewRequest(http.MethodPost, "/api/sendCOin", bytes.NewReader([]byte("invalid")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.Set("user", senderToken)
		err := handlerCombined.SendCoinHandler(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid request")
	})
}
