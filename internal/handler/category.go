package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
	validator       *validator.Validator
}

func NewCategoryHandler(categoryService *service.CategoryService, v *validator.Validator) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService, validator: v}
}

func (h *CategoryHandler) FindAll(c *gin.Context) {
	categories, err := h.categoryService.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to retrieve categories")
		return
	}
	response.OK(c, "Success", categories)
}

func (h *CategoryHandler) FindByID(c *gin.Context) {
	id := c.Param("id")
	category, err := h.categoryService.FindByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "Success", category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req entity.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Category created successfully", category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req entity.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Category updated successfully", category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.categoryService.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Category deleted successfully", nil)
}
