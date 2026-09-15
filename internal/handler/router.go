package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/config"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/middleware"
	"github.com/terenjit/Cafe-POS/internal/repository"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

func NewRouter(db *sql.DB, cfg *config.Config, v *validator.Validator) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	authHandler := NewAuthHandler(authService, v)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.OK(c, "server is running", nil)
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		ownerGroup := v1.Group("/owner")
		ownerGroup.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		ownerGroup.Use(middleware.RoleMiddleware(entity.RoleOwner))
		_ = ownerGroup

		cashierGroup := v1.Group("/cashier")
		cashierGroup.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		cashierGroup.Use(middleware.RoleMiddleware(entity.RoleCashier))
		_ = cashierGroup
	}

	return r
}
