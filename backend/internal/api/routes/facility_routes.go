package routes

import (
	"github.com/fasim/backend/internal/api/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterFacilityRoutes registers all facility-related routes
func RegisterFacilityRoutes(e *echo.Echo, handler *handlers.FacilityHandler) {
	facilities := e.Group("/api/facilities")
	facilities.GET("", handler.List)
	facilities.GET("/:id", handler.Get)
	facilities.POST("", handler.Create)
	facilities.PUT("/:id", handler.Update)
	facilities.DELETE("/:id", handler.Delete)
}
