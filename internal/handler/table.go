package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/service"
	"github.com/terenjit/Cafe-POS/pkg/response"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

type TableHandler struct {
	tableService *service.TableService
	validator    *validator.Validator
}

func NewTableHandler(tableService *service.TableService, v *validator.Validator) *TableHandler {
	return &TableHandler{tableService: tableService, validator: v}
}

func (h *TableHandler) FindAll(c *gin.Context) {
	tables, err := h.tableService.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to retrieve tables")
		return
	}
	response.OK(c, "Success", tables)
}

func (h *TableHandler) Create(c *gin.Context) {
	var req entity.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	table, err := h.tableService.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Table created successfully", table)
}

func (h *TableHandler) Update(c *gin.Context) {
	var req entity.UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	table, err := h.tableService.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Table updated successfully", table)
}

func (h *TableHandler) Delete(c *gin.Context) {
	if err := h.tableService.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "Table deleted successfully", nil)
}
