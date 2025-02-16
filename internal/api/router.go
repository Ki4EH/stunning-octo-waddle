package api

import (
	"github.com/Ki4EH/stunning-octo-waddle/internal/api/handler"
	"github.com/Ki4EH/stunning-octo-waddle/internal/api/middleware"
	"github.com/Ki4EH/stunning-octo-waddle/internal/db/repository"
	"github.com/Ki4EH/stunning-octo-waddle/internal/logger"
	"github.com/Ki4EH/stunning-octo-waddle/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	middlewareEcho "github.com/labstack/echo/v4/middleware"
)

func InitRoutes(e *echo.Echo, db *pgxpool.Pool, log *logger.Logger) {
	e.Use(middleware.LoggingMiddleware(*log))
	e.Use(middlewareEcho.Recover())
	e.Use(middleware.CORSConfig())

	userRepo := repository.NewUserRepository(db)
	authHandler := handler.NewAuthorizationHandler(userRepo)

	// Путь для авторизации
	e.POST("/api/auth", authHandler.Login)

	apiGroup := e.Group("/api")

	apiGroup.Use(echojwt.WithConfig(utils.JwtConfig))

	coinRepo := repository.NewCoinRepository(db)
	coinHandler := handler.NewCoinHandler(coinRepo)

	combinedRepository := handler.NewCombinedRepository(userRepo, coinRepo)

	apiGroup.GET("/info", combinedRepository.GetInfo)

	apiGroup.GET("/buy/:item", coinHandler.BuyItem)
	apiGroup.POST("/sendCoin", combinedRepository.SendCoinHandler)

}
