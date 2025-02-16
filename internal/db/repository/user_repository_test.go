package repository

import (
	"context"
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/config"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	mock_repository "github.com/Ki4EH/stunning-octo-waddle/internal/mocks/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func TestInterfaceGetUserCredentialByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockUserRepository(ctrl)
	ctx := context.Background()
	expectedCredential := &models.Credential{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "testpassword",
	}

	mockRepo.EXPECT().
		GetUserCredentialByName(ctx, "testuser").
		Return(expectedCredential, nil)

	credential, err := mockRepo.GetUserCredentialByName(ctx, "testuser")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if credential != expectedCredential {
		t.Fatalf("expected credential %v, got %v", expectedCredential, credential)
	}
}

func TestInterfaceGetUserItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockUserRepository(ctrl)
	ctx := context.Background()
	userID := uuid.New()

	expectedItems := []models.UserItem{
		{Type: "sword", Quantity: 1},
		{Type: "shield", Quantity: 2},
	}

	mockRepo.EXPECT().
		GetUserItems(ctx, userID, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ uuid.UUID, items *[]models.UserItem) error {
			*items = append(*items, expectedItems...)
			return nil
		})

	var actualItems []models.UserItem
	err := mockRepo.GetUserItems(ctx, userID, &actualItems)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(actualItems) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(actualItems))
	}
	for i, item := range actualItems {
		if item != expectedItems[i] {
			t.Fatalf("Item mismatch at index %d: expected %v, got %v", i, expectedItems[i], item)
		}
	}
}

func TestInterfaceGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockUserRepository(ctrl)
	ctx := context.Background()
	userID := uuid.New()

	expectedCredential := &models.Credential{
		ID:       userID,
		Username: "test_user",
		Coin:     500,
	}

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(expectedCredential, nil)

	credential, err := mockRepo.GetUserByID(ctx, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if credential.ID != userID {
		t.Fatalf("Expected user ID %v, got %v", userID, credential.ID)
	}
	if credential.Username != expectedCredential.Username {
		t.Fatalf("Expected username %s, got %s", expectedCredential.Username, credential.Username)
	}
	if credential.Coin != expectedCredential.Coin {
		t.Fatalf("Expected coin balance %d, got %d", expectedCredential.Coin, credential.Coin)
	}
}

func TestInterfaceCreateUserCredential(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_repository.NewMockUserRepository(ctrl)
	ctx := context.Background()

	newCredential := &models.Credential{
		ID:       uuid.New(),
		Username: "new_user",
		Password: "secure_password",
	}

	mockRepo.EXPECT().
		CreateUserCredential(ctx, newCredential).
		Return(nil)

	err := mockRepo.CreateUserCredential(ctx, newCredential)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	cfg, _ := config.LoadConfig()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)

	ctx := context.Background()

	var err error
	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	os.Exit(m.Run())
}

func setupUser() (repo *userRepository, ctx context.Context) {
	ctx = context.Background()

	repo = &userRepository{db: pool}

	return repo, ctx
}

func TestGetUserCredentialByName(t *testing.T) {
	repo, ctx := setupUser()

	tx, err := repo.db.Begin(ctx)
	require.NoError(t, err)
	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			t.Fatalf("failed to rollback transaction: %v", err)
		}
	}(tx, ctx)

	t.Run("existing user", func(t *testing.T) {
		expected := models.Credential{
			ID:       uuid.New(),
			Username: uuid.NewString(),
			Password: "TestGetUserCredentialByName",
		}
		_, err := repo.db.Exec(ctx, `
      INSERT INTO credentials (username, password)
      VALUES ($1, $2)
    `, expected.Username, expected.Password)
		require.NoError(t, err)

		result, err := repo.GetUserCredentialByName(ctx, expected.Username)
		require.NoError(t, err)
		require.Equal(t, expected.Username, result.Username)
	})

	t.Run("non-existing user", func(t *testing.T) {
		result, err := repo.GetUserCredentialByName(ctx, "non_existent")
		require.NoError(t, err)
		require.Empty(t, result.Username)
	})
}

func TestCreateUserCredential(t *testing.T) {
	repo, ctx := setupUser()

	userName := uuid.NewString()

	cred := &models.Credential{
		ID:       uuid.New(),
		Username: userName,
		Password: "TestCreateUserCredential",
	}

	err := repo.CreateUserCredential(ctx, cred)
	require.NoError(t, err)

	var password string
	err = repo.db.QueryRow(ctx, `
        SELECT password FROM credentials WHERE username = $1
    `, cred.Username).Scan(&password)
	require.NoError(t, err)
	require.Equal(t, cred.Password, password)
}

func TestGetUserByID(t *testing.T) {
	repo, ctx := setupUser()

	t.Run("existing user", func(t *testing.T) {
		expected := models.Credential{
			ID:       uuid.New(),
			Username: uuid.NewString(),
			Password: "TestGetUserByID",
			Coin:     100,
		}
		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, coin, password)
            VALUES ($1, $2, $3, 'TestGetUserByID')
        `, expected.ID, expected.Username, expected.Coin)
		require.NoError(t, err)

		result, err := repo.GetUserByID(ctx, expected.ID)
		require.NoError(t, err)
		require.Equal(t, expected.Username, result.Username)
		require.Equal(t, expected.Coin, result.Coin)
	})

	t.Run("non-existing user", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, uuid.New())
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetUserItems(t *testing.T) {
	repo, ctx := setupUser()

	userID := uuid.New()
	_, err := repo.db.Exec(ctx, `
        INSERT INTO credentials (id, username, password)
        VALUES ($1, $2, 'TestGetUserItems')
    `, userID, userID)
	require.NoError(t, err)

	expectedItems := []models.UserItem{
		{Type: "cup", Quantity: 1},
		{Type: "pen", Quantity: 2},
	}

	for _, item := range expectedItems {
		_, err = repo.db.Exec(ctx, `
            INSERT INTO user_items (user_id, type, quantity)
            VALUES ($1, $2, $3)
        `, userID, item.Type, item.Quantity)
		require.NoError(t, err)
	}

	var items []models.UserItem
	err = repo.GetUserItems(ctx, userID, &items)
	require.NoError(t, err)
	require.Len(t, items, len(expectedItems))
	require.ElementsMatch(t, expectedItems, items)
}

func TestGetUsernamesByIDs(t *testing.T) {
	repo, ctx := setupUser()

	users := map[string]string{
		uuid.New().String(): uuid.NewString(),
		uuid.New().String(): uuid.NewString(),
	}

	for id, username := range users {
		_, err := repo.db.Exec(ctx, `
            INSERT INTO credentials (id, username, password)
            VALUES ($1, $2, 'TestGetUsernamesByIDs')
        `, uuid.MustParse(id), username)
		require.NoError(t, err)
	}

	userIDs := make([]string, 0, len(users))
	for id := range users {
		userIDs = append(userIDs, id)
	}

	result, err := repo.GetUsernamesByIDs(ctx, userIDs)
	require.NoError(t, err)
	require.Len(t, result, len(users))

	for id, expectedUsername := range users {
		require.Equal(t, expectedUsername, result[id])
	}

	t.Run("empty input", func(t *testing.T) {
		result, err := repo.GetUsernamesByIDs(ctx, []string{})
		require.NoError(t, err)
		require.Nil(t, result)
	})
}
