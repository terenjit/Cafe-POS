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

type OrderHandler struct {
	orderService *service.OrderService
	validator    *validator.Validator
}

func NewOrderHandler(orderService *service.OrderService, v *validator.Validator) *OrderHandler {
	return &OrderHandler{orderService: orderService, validator: v}
}

func (h *OrderHandler) FindOrders(c *gin.Context) {
	cashierID := c.GetString("user_id")

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	filter := repository.OrderFilter{
		Status: c.Query("status"),
		Page:   page,
		Limit:  limit,
	}

	orders, total, err := h.orderService.FindOrders(c.Request.Context(), cashierID, filter)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Success", PaginatedResponse{
		Items: orders,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req entity.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	order, err := h.orderService.Create(c.Request.Context(), cashierID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Order created successfully", order)
}

func (h *OrderHandler) AddItem(c *gin.Context) {
	var req entity.AddOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	order, err := h.orderService.AddItem(c.Request.Context(), cashierID, orderID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Item added successfully", order)
}

func (h *OrderHandler) UpdateItem(c *gin.Context) {
	var req entity.UpdateOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	itemID := c.Param("item_id")
	order, err := h.orderService.UpdateItem(c.Request.Context(), cashierID, orderID, itemID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Item updated successfully", order)
}

func (h *OrderHandler) RemoveItem(c *gin.Context) {
	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	itemID := c.Param("item_id")
	order, err := h.orderService.RemoveItem(c.Request.Context(), cashierID, orderID, itemID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Item removed successfully", order)
}

func (h *OrderHandler) ApplyPromo(c *gin.Context) {
	var req entity.ApplyPromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	order, err := h.orderService.ApplyPromo(c.Request.Context(), cashierID, orderID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Promo applied successfully", order)
}

func (h *OrderHandler) RemovePromo(c *gin.Context) {
	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	order, err := h.orderService.RemovePromo(c.Request.Context(), cashierID, orderID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Promo removed successfully", order)
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	var req entity.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	orderID := c.Param("id")
	checkoutResponse, err := h.orderService.Checkout(c.Request.Context(), cashierID, orderID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Checkout successful", checkoutResponse)
}

func (h *OrderHandler) FindByID(c *gin.Context) {
	order, err := h.orderService.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", order)
}
