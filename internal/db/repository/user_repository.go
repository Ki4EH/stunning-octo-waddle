package repository

import (
	"context"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetUserCredentialByName(ctx context.Context, name string) (*models.Credential, error)
	CreateUserCredential(ctx context.Context, credential *models.Credential) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.Credential, error)
	GetUserItems(ctx context.Context, id uuid.UUID, userItems *[]models.UserItem) error
	GetUsernamesByIDs(ctx context.Context, userIDs []string) (map[string]string, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetUserCredentialByName(ctx context.Context, name string) (*models.Credential, error) {
	var credential models.Credential
	row := r.db.QueryRow(ctx, "SELECT id, username, password FROM credentials WHERE username = $1", name)
	err := row.Scan(&credential.ID, &credential.Username, &credential.Password)
	if err != nil && err.Error() != "no rows in result set" {
		return nil, err
	}
	return &credential, nil
}

func (r *userRepository) CreateUserCredential(ctx context.Context, credential *models.Credential) error {
	err := r.db.QueryRow(ctx, "INSERT INTO credentials (username, password) VALUES ($1, $2) RETURNING id", credential.Username, credential.Password).Scan(&credential.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.Credential, error) {
	var credential models.Credential
	row := r.db.QueryRow(ctx, "SELECT id, username, coin FROM credentials WHERE id = $1", id)
	err := row.Scan(&credential.ID, &credential.Username, &credential.Coin)
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

func (r *userRepository) GetUserItems(ctx context.Context, id uuid.UUID, userItems *[]models.UserItem) error {
	rows, err := r.db.Query(ctx, "SELECT type, quantity FROM user_items WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userItem models.UserItem
		err = rows.Scan(&userItem.Type, &userItem.Quantity)
		if err != nil {
			return err
		}
		*userItems = append(*userItems, userItem)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUsernamesByIDs(ctx context.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	query := `
        SELECT id, username
        FROM credentials
        WHERE id IN (SELECT unnest($1::uuid[]))`

	rows, err := r.db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usernameMap := make(map[string]string)
	for rows.Next() {
		var id, username string
		if err = rows.Scan(&id, &username); err != nil {
			return nil, err
		}
		usernameMap[id] = username
	}
	return usernameMap, nil
}
