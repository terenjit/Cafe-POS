package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type PromoHandler struct {
	promoService *service.PromoService
	validator    *validator.Validator
}

func NewPromoHandler(promoService *service.PromoService, v *validator.Validator) *PromoHandler {
	return &PromoHandler{promoService: promoService, validator: v}
}

func (h *PromoHandler) FindAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	promos, total, err := h.promoService.FindAll(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "Failed to retrieve promos")
		return
	}
	response.OK(c, "Success", PaginatedResponse{
		Items: promos,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *PromoHandler) Create(c *gin.Context) {
	var req entity.CreatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	promo, err := h.promoService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Promo created successfully", promo)
}

func (h *PromoHandler) Update(c *gin.Context) {
	var req entity.UpdatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	promo, err := h.promoService.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Promo updated successfully", promo)
}

func (h *PromoHandler) Delete(c *gin.Context) {
	if err := h.promoService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Promo deleted successfully", nil)
}
