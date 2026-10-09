package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type ShiftHandler struct {
	shiftService *service.ShiftService
	validator    *validator.Validator
}

func NewShiftHandler(shiftService *service.ShiftService, v *validator.Validator) *ShiftHandler {
	return &ShiftHandler{shiftService: shiftService, validator: v}
}

func (h *ShiftHandler) Open(c *gin.Context) {
	var req entity.OpenShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	shift, err := h.shiftService.Open(c.Request.Context(), cashierID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Shift opened successfully", shift)
}

func (h *ShiftHandler) Close(c *gin.Context) {
	var req entity.CloseShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	cashierID := c.GetString("user_id")
	shift, err := h.shiftService.Close(c.Request.Context(), cashierID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Shift closed successfully", shift)
}

func (h *ShiftHandler) GetCurrent(c *gin.Context) {
	cashierID := c.GetString("user_id")
	shift, err := h.shiftService.GetCurrent(c.Request.Context(), cashierID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", shift)
}
