package handler

import (
	"errors"
	mock_repository "github.com/Ki4EH/stunning-octo-waddle/internal/mocks/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoinHandler_BuyItem(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		setupMocks     func(*mock_repository.MockCoinRepository)
		token          *jwt.Token
		item           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful purchase",
			setupMocks: func(mockRepo *mock_repository.MockCoinRepository) {
				mockRepo.EXPECT().
					BuyItemFromShop(gomock.Any(), gomock.Any(), "item1").
					Return(nil)
			},
			token: &jwt.Token{
				Claims: &utils.Claims{
					UserID: uuid.New(),
				},
			},
			item:           "item1",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name: "failed to get JWT token",
			setupMocks: func(mockRepo *mock_repository.MockCoinRepository) {
			},
			token:          nil, // симулируем отсутствие токена
			item:           "item1",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"errors":"failed to get jwt token"}`,
		},
		{
			name: "failed to buy item - insufficient funds",
			setupMocks: func(mockRepo *mock_repository.MockCoinRepository) {
				mockRepo.EXPECT().
					BuyItemFromShop(gomock.Any(), gomock.Any(), "item1").
					Return(errors.New("insufficient funds"))
			},
			token: &jwt.Token{
				Claims: &utils.Claims{
					UserID: uuid.New(),
				},
			},
			item:           "item1",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errors":"failed to buy item insufficient funds"}`,
		},
		{
			name: "failed to buy item - repository error",
			setupMocks: func(mockRepo *mock_repository.MockCoinRepository) {
				mockRepo.EXPECT().
					BuyItemFromShop(gomock.Any(), gomock.Any(), "item1").
					Return(errors.New("database error"))
			},
			token: &jwt.Token{
				Claims: &utils.Claims{
					UserID: uuid.New(),
				},
			},
			item:           "item1",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errors":"failed to buy item database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockCoinRepository(ctrl)

			tt.setupMocks(mockRepo)

			handler := &CoinHandler{
				repo: mockRepo,
			}

			req := httptest.NewRequest(http.MethodPost, "/buy/item1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("item")
			c.SetParamValues(tt.item)

			if tt.token != nil {
				c.Set("user", tt.token)
			}

			err := handler.BuyItem(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.name == "successful purchase" {
				assert.Empty(t, rec.Body.String())
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
