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

type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

type ProductHandler struct {
	productService *service.ProductService
	validator      *validator.Validator
}

func NewProductHandler(productService *service.ProductService, v *validator.Validator) *ProductHandler {
	return &ProductHandler{productService: productService, validator: v}
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	search := c.Query("search")
	categoryID := c.Query("category_id")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	var isActive *bool
	if raw := c.Query("is_active"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			isActive = &parsed
		}
	}

	filter := repository.ProductFilter{
		Search:     search,
		CategoryID: categoryID,
		IsActive:   isActive,
		Page:       page,
		Limit:      limit,
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	products, total, err := h.productService.FindAll(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "Failed to retrieve products")
		return
	}

	response.OK(c, "Success", PaginatedResponse{
		Items: products,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	})
}

func (h *ProductHandler) FindByID(c *gin.Context) {
	product, err := h.productService.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req entity.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	product, err := h.productService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Product created successfully", product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var req entity.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	product, err := h.productService.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Product updated successfully", product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	if err := h.productService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Product deleted successfully", nil)
}
