package e2e_test

import (
	"context"
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/api/handler"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	testDB *pgxpool.Pool
)

// TestMain подготавливает контейнер и базу данных перед запуском e2e тестов
func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, dbPool, err := setupPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to start container: %v", err))
	}
	defer func() {
		if err = pgContainer.Terminate(ctx); err != nil {
			panic(fmt.Sprintf("Failed to terminate container: %v", err))
		}
	}()

	testDB = dbPool

	// Тут мы инициализируем тестовую базу данных
	err = initTestDB(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize test database: %v", err))
	}

	os.Exit(m.Run())
}

func setupPostgresContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, error) {
	timeout := 120 * time.Second
	ctxT, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // Тут надо проверить что регает
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(timeout),
	}

	container, err := testcontainers.GenericContainer(ctxT,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctxT)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctxT, "5432")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get port: %w", err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"testuser",
		"testpass",
		host,
		port.Port(),
		"testdb",
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse config: %w", err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return container, dbPool, nil
}

func initTestDB(ctx context.Context) error {
	_, err := testDB.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
        CREATE TABLE IF NOT EXISTS credentials (
            id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4() NOT NULL,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            coin BIGINT DEFAULT 0
        );
        CREATE TABLE IF NOT EXISTS shops (
            item TEXT PRIMARY KEY,
            price BIGINT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS user_items (
            user_id UUID REFERENCES credentials(id),
            type TEXT,
            quantity INT,
            PRIMARY KEY (user_id, type)
        );
		CREATE TABLE IF NOT EXISTS credentials (
			id UUID PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			coin BIGINT DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT public.uuid_generate_v4() NOT NULL,
			from_user UUID,
			to_user UUID,
			amount BIGINT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS shops (
			item TEXT PRIMARY KEY,
			price BIGINT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS user_items (
			user_id UUID REFERENCES credentials(id),
		type TEXT,
		quantity INT,
			PRIMARY KEY (user_id, type)
	);
`)
	return err
}

func setupCoinTest(t *testing.T) (context.Context, *handler.CoinHandler) {
	ctx := context.Background()

	// Очистка тестовых данных
	_, err := testDB.Exec(ctx, `
        TRUNCATE credentials, shops, user_items, transactions CASCADE;
    `)
	require.NoError(t, err)

	// Добавление кружки в магазин для тестов
	_, err = testDB.Exec(ctx, `
        INSERT INTO shops (item, price)
        VALUES ('cup', 20);
    `)
	require.NoError(t, err)

	repo := repository.NewCoinRepository(testDB)
	handlerCoin := handler.NewCoinHandler(repo)

	return ctx, handlerCoin
}

func createTestUser(t *testing.T, coin int64) (uuid.UUID, *jwt.Token) {
	ctx := context.Background()
	userID := uuid.New()

	_, err := testDB.Exec(ctx, `
        INSERT INTO credentials (id, username, password, coin)
        VALUES ($1, $2, $3, $4)
    `, userID, userID.String(), "testpass", coin)
	require.NoError(t, err)

	claims := &utils.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // срок жизни токена
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return userID, token
}

func TestBuyItemHandler(t *testing.T) {
	t.Run("successful purchase", func(t *testing.T) {
		ctx, handlerCoin := setupCoinTest(t)

		// Создаю тестового пользователя с балансом 20 монет
		userID, token := createTestUser(t, 20)

		e := echo.New()

		e.Use(echojwt.WithConfig(utils.JwtConfig))

		req := httptest.NewRequest(http.MethodPost, "/api/buy/cup", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/api/buy/:item")
		c.SetParamNames("item")
		c.SetParamValues("cup")

		c.Set("user", token)
		err := handlerCoin.BuyItem(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var balance int64
		err = testDB.QueryRow(ctx,
			"SELECT coin FROM credentials WHERE id = $1", userID).
			Scan(&balance)
		require.NoError(t, err)
		require.Equal(t, int64(0), balance)

		var quantity int
		err = testDB.QueryRow(ctx, `
            SELECT quantity FROM user_items 
            WHERE user_id = $1 AND type = 'cup'
        `, userID).Scan(&quantity)
		require.NoError(t, err)
		require.Equal(t, 1, quantity)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		_, handlerCoin := setupCoinTest(t)

		_, token := createTestUser(t, 5)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/buy/cup", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/buy/:item")
		c.SetParamNames("item")
		c.SetParamValues("cup")
		c.Set("user", token)

		err := handlerCoin.BuyItem(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "insufficient balance")
	})

	t.Run("invalid item", func(t *testing.T) {
		_, handlerCoin := setupCoinTest(t)

		_, token := createTestUser(t, 1000)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/buy/invalid", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/buy/:item")
		c.SetParamNames("item")
		c.SetParamValues("invalid")
		c.Set("user", token)

		err := handlerCoin.BuyItem(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "item not found")
	})

	t.Run("invalid token", func(t *testing.T) {
		_, handlerCoin := setupCoinTest(t)

		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/api/buy/cup", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/buy/:item")
		c.SetParamNames("item")
		c.SetParamValues("cup")

		err := handlerCoin.BuyItem(c)

		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Contains(t, rec.Body.String(), "failed to get jwt token")
	})
}
