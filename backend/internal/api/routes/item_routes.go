package routes

import (
	"github.com/fasim/backend/internal/api/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterItemRoutes registers all item-related routes
func RegisterItemRoutes(e *echo.Echo, handler *handlers.ItemHandler) {
	items := e.Group("/api/items")
	items.GET("", handler.List)
	items.GET("/:id", handler.Get)
	items.POST("", handler.Create)
	items.PUT("/:id", handler.Update)
	items.DELETE("/:id", handler.Delete)
}
