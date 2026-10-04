package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type UserHandler struct {
	userService *service.UserService
	validator   *validator.Validator
}

func NewUserHandler(userService *service.UserService, v *validator.Validator) *UserHandler {
	return &UserHandler{userService: userService, validator: v}
}

func (h *UserHandler) FindAll(c *gin.Context) {
	cashiers, err := h.userService.FindAllCashiers(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to retrieve cashiers")
		return
	}
	response.OK(c, "Success", cashiers)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req entity.CreateCashierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.userService.CreateCashier(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Cashier created successfully", result)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req entity.UpdateCashierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.userService.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Cashier updated successfully", result)
}

func (h *UserHandler) ToggleStatus(c *gin.Context) {
	result, err := h.userService.ToggleStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Cashier status updated successfully", result)
}
