package handler

import (
	"fmt"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/models"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

type AuthorizationHandler struct {
	repo repository.UserRepository
}

func NewAuthorizationHandler(repo repository.UserRepository) *AuthorizationHandler {
	return &AuthorizationHandler{
		repo: repo,
	}
}

func (r *AuthorizationHandler) Login(c echo.Context) error {
	var loginRequest models.Credential
	if err := c.Bind(&loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"errors": fmt.Sprintf("invalid request %v", err)})
	}
	if loginRequest.Username == "" || loginRequest.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"errors": "username and password are required"})
	}

	credential, err := r.repo.GetUserCredentialByName(c.Request().Context(), loginRequest.Username)
	if err != nil {
		c.Logger().Error("failed to fetch user info", err)
		c.Response().Status = http.StatusInternalServerError
		return c.JSON(http.StatusUnauthorized, map[string]string{"errors": fmt.Sprintf("failed to fetch user info %v", err)})
	}

	if credential.ID == uuid.Nil {
		credential.Username = loginRequest.Username
		credential.Password = loginRequest.Password
		err = r.repo.CreateUserCredential(c.Request().Context(), credential)
		if err != nil {
			c.Response().Status = http.StatusInternalServerError
			c.Logger().Error("failed to create user", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"errors": fmt.Sprintf("failed to create user %v", err)})
		}
	} else {
		if credential.Password != loginRequest.Password {
			c.Response().Status = http.StatusUnauthorized
			return c.JSON(http.StatusUnauthorized, map[string]string{"errors": "invalid password"})
		}
	}

	token, err := utils.GenerateToken(credential.ID)
	if err != nil {
		c.Response().Status = http.StatusInternalServerError
		c.Logger().Error("failed to generate token", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"errors": fmt.Sprintf("failed to generate token %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}
