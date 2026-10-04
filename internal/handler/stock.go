package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/repository"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type StockHandler struct {
	stockService *service.StockService
	validator    *validator.Validator
}

func NewStockHandler(stockService *service.StockService, v *validator.Validator) *StockHandler {
	return &StockHandler{stockService: stockService, validator: v}
}

func (h *StockHandler) GetMovements(c *gin.Context) {
	productID := c.Param("id")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	typeFilter := c.Query("type")

	filter := repository.StockFilter{Page: page, Limit: limit, Type: typeFilter}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	movements, total, err := h.stockService.GetMovements(c.Request.Context(), productID, filter)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", PaginatedResponse{
		Items: movements,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	})
}

func (h *StockHandler) GetStock(c *gin.Context) {
	product, err := h.stockService.GetStock(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", gin.H{"stock": product.Stock, "product_id": product.ID, "name": product.Name})
}

func (h *StockHandler) Adjust(c *gin.Context) {
	productID := c.Param("id")
	userID := c.GetString("user_id")

	var req entity.StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	if err := h.stockService.Adjust(c.Request.Context(), productID, userID, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Stock updated successfully", nil)
}
