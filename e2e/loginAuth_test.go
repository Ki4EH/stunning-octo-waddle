package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Ki4EH/stunning-octo-waddle/internal/api/handler"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupUserTest(t *testing.T) (context.Context, *handler.AuthorizationHandler) {
	ctx := context.Background()

	_, err := testDB.Exec(ctx, `
        TRUNCATE credentials, transactions CASCADE;
    `)
	require.NoError(t, err)

	userRepo := repository.NewUserRepository(testDB)
	authorization := handler.NewAuthorizationHandler(userRepo)

	return ctx, authorization
}

func TestLoginHandler(t *testing.T) {
	t.Run("successful user registration", func(t *testing.T) {
		_, handlerAuth := setupUserTest(t)

		e := echo.New()

		loginReq := models.Credential{
			Username: "newuser",
			Password: "securepassword123",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handlerAuth.Login(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var token struct{ Token string }
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &token))
		require.NotEmpty(t, token)

		var count int
		err = testDB.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM credentials WHERE username = $1", "newuser").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("successful login with existing user", func(t *testing.T) {
		ctx, handlerAuth := setupUserTest(t)

		password := "validpassword"

		_, err := testDB.Exec(ctx, `
            INSERT INTO credentials (username, password)
            VALUES ($1, $2)
        `, "existinguser", password)
		require.NoError(t, err)

		loginReq := models.Credential{
			Username: "existinguser",
			Password: password,
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		err = handlerAuth.Login(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var token struct{ Token string }
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &token))
		claims, err := utils.ParseToken(token.Token)
		require.NoError(t, err)
		require.NotEmpty(t, claims.UserID)
	})

	t.Run("invalid password", func(t *testing.T) {
		ctx, handlerAuth := setupUserTest(t)

		_, err := testDB.Exec(ctx, `
            INSERT INTO credentials (username, password)
            VALUES ($1, $2)
        `, "existinguser", "correctpassword")
		require.NoError(t, err)

		loginReq := models.Credential{
			Username: "existinguser",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		err = handlerAuth.Login(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid password")
	})

	t.Run("invalid request", func(t *testing.T) {
		_, handlerAuth := setupUserTest(t)

		body := []byte(`{"invalid": "request"`)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		err := handlerAuth.Login(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid request")
	})

	t.Run("duplicate username registration", func(t *testing.T) {
		ctx, handlerAuth := setupUserTest(t)

		_, err := testDB.Exec(ctx, `
            INSERT INTO credentials (username, password)
            VALUES ($1, $2)
        `, "duplicateuser", "password")
		require.NoError(t, err)

		loginReq := models.Credential{
			Username: "duplicateuser",
			Password: "newpassword",
		}
		body, _ := json.Marshal(loginReq)

		req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		err = handlerAuth.Login(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Contains(t, rec.Body.String(), "invalid password")
	})
}
