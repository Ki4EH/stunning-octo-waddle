package handler

import (
	"bytes"
	"errors"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/mocks/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizationHandler_Login(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		setupMocks     func(*mock_repository.MockUserRepository)
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "invalid request body",
			setupMocks: func(mockRepo *mock_repository.MockUserRepository) {
			},
			requestBody:    `{"username":"testuser"`, // Ошибка в JSON
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"errors":"invalid request code=400, message=unexpected EOF, internal=unexpected EOF"}`,
		},
		{
			name: "failed to fetch user info",
			setupMocks: func(mockRepo *mock_repository.MockUserRepository) {
				mockRepo.EXPECT().
					GetUserCredentialByName(gomock.Any(), "testuser").
					Return(nil, errors.New("database error"))
			},
			requestBody:    `{"username":"testuser","password":"testpass"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errors":"failed to fetch user info database error"}`,
		},
		{
			name: "invalid password",
			setupMocks: func(mockRepo *mock_repository.MockUserRepository) {
				mockRepo.EXPECT().
					GetUserCredentialByName(gomock.Any(), "testuser").
					Return(&models.Credential{
						ID:       uuid.New(),
						Username: "testuser",
						Password: "hashedpassword",
					}, nil)
			},
			requestBody:    `{"username":"testuser","password":"wrongpass"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"errors": "invalid password"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repository.NewMockUserRepository(ctrl)

			tt.setupMocks(mockRepo)

			handler := &AuthorizationHandler{
				repo: mockRepo,
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Login(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
		})
	}
}
