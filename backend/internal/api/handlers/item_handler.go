package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

type ItemHandler struct {
	repo repositories.ItemRepository
}

func NewItemHandler(repo repositories.ItemRepository) *ItemHandler {
	return &ItemHandler{repo: repo}
}

type createItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type itemResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func toItemResponse(item *models.Item) itemResponse {
	return itemResponse{
		ID:          item.ID(),
		Name:        item.Name(),
		Description: item.Description(),
	}
}

// List handles GET /api/items
func (h *ItemHandler) List(c echo.Context) error {
	items, err := h.repo.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responses := make([]itemResponse, len(items))
	for i, item := range items {
		responses[i] = toItemResponse(item)
	}

	return c.JSON(http.StatusOK, responses)
}

// Get handles GET /api/items/:id
func (h *ItemHandler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item ID")
	}

	item, err := h.repo.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Item not found")
	}

	return c.JSON(http.StatusOK, toItemResponse(item))
}

// Create handles POST /api/items
func (h *ItemHandler) Create(c echo.Context) error {
	var req createItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	fmt.Printf("Request: name=%s, description=%s\n", req.Name, req.Description)
	item := models.NewItem(req.Name, req.Description)
	fmt.Printf("Created item: name=%s, description=%s\n", item.Name(), item.Description())
	if err := h.repo.Create(c.Request().Context(), item); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, toItemResponse(item))
}

// Update handles PUT /api/items/:id
func (h *ItemHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item ID")
	}

	var req updateItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	existingItem, err := h.repo.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if existingItem == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Item not found")
	}

	updatedItem := models.NewItemFromParams(id, req.Name, req.Description)
	if err := h.repo.Update(c.Request().Context(), updatedItem); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, toItemResponse(updatedItem))
}

// Delete handles DELETE /api/items/:id
func (h *ItemHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid item ID")
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
