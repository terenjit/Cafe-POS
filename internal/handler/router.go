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
	"github.com/terenjit/Cafe-POS/pkg/txmanager"
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
	userService := service.NewUserService(userRepo, authService)
	userHandler := NewUserHandler(userService, v)

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := NewCategoryHandler(categoryService, v)

	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo, categoryRepo)
	productHandler := NewProductHandler(productService, v)

	stockRepo := repository.NewStockRepository(db)
	txMgr := txmanager.New(db)
	stockService := service.NewStockService(stockRepo, productRepo, txMgr)
	stockHandler := NewStockHandler(stockService, v)

	tableRepo := repository.NewTableRepository(db)
	tableService := service.NewTableService(tableRepo)
	tableHandler := NewTableHandler(tableService, v)

	promoRepo := repository.NewPromoRepository(db)
	promoService := service.NewPromoService(promoRepo)
	promoHandler := NewPromoHandler(promoService, v)

	orderRepo := repository.NewOrderRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	shiftService := service.NewShiftService(shiftRepo, orderRepo)
	shiftHandler := NewShiftHandler(shiftService, v)

	paymentRepo := repository.NewPaymentRepository(db)
	orderService := service.NewOrderService(orderRepo, shiftRepo, tableRepo, productRepo, promoRepo, paymentRepo, txMgr)
	orderHandler := NewOrderHandler(orderService, v)

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

		ownerGroup.GET("/categories", categoryHandler.FindAll)
		ownerGroup.GET("/categories/:id", categoryHandler.FindByID)
		ownerGroup.POST("/categories", categoryHandler.Create)
		ownerGroup.PUT("/categories/:id", categoryHandler.Update)
		ownerGroup.DELETE("/categories/:id", categoryHandler.Delete)

		ownerGroup.GET("/products", productHandler.FindAll)
		ownerGroup.GET("/products/:id", productHandler.FindByID)
		ownerGroup.POST("/products", productHandler.Create)
		ownerGroup.PUT("/products/:id", productHandler.Update)
		ownerGroup.DELETE("/products/:id", productHandler.Delete)

		ownerGroup.GET("/products/:id/stock", stockHandler.GetStock)
		ownerGroup.GET("/products/:id/stock/movements", stockHandler.GetMovements)
		ownerGroup.POST("/products/:id/stock/adjustment", stockHandler.Adjust)

		ownerGroup.GET("/cashiers", userHandler.FindAll)
		ownerGroup.POST("/cashiers", userHandler.Create)
		ownerGroup.PUT("/cashiers/:id", userHandler.Update)
		ownerGroup.PATCH("/cashiers/:id/toggle-status", userHandler.ToggleStatus)

		ownerGroup.GET("/tables", tableHandler.FindAll)
		ownerGroup.POST("/tables", tableHandler.Create)
		ownerGroup.PUT("/tables/:id", tableHandler.Update)
		ownerGroup.DELETE("/tables/:id", tableHandler.Delete)

		ownerGroup.GET("/promos", promoHandler.FindAll)
		ownerGroup.POST("/promos", promoHandler.Create)
		ownerGroup.PUT("/promos/:id", promoHandler.Update)
		ownerGroup.DELETE("/promos/:id", promoHandler.Delete)

		cashierGroup := v1.Group("/cashier")
		cashierGroup.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		cashierGroup.Use(middleware.RoleMiddleware(entity.RoleCashier))

		cashierGroup.POST("/shifts/open", shiftHandler.Open)
		cashierGroup.POST("/shifts/close", shiftHandler.Close)
		cashierGroup.GET("/shifts/current", shiftHandler.GetCurrent)

		cashierGroup.GET("/orders/history", orderHandler.FindOrders)
		cashierGroup.POST("/orders", orderHandler.Create)
		cashierGroup.GET("/orders/:id", orderHandler.FindByID)
		cashierGroup.POST("/orders/:id/items", orderHandler.AddItem)
		cashierGroup.PUT("/orders/:id/items/:item_id", orderHandler.UpdateItem)
		cashierGroup.DELETE("/orders/:id/items/:item_id", orderHandler.RemoveItem)
		cashierGroup.POST("/orders/:id/promo", orderHandler.ApplyPromo)
		cashierGroup.DELETE("/orders/:id/promo", orderHandler.RemovePromo)
		cashierGroup.POST("/orders/:id/checkout", orderHandler.Checkout)
	}

	return r
}
