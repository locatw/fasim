package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasim/backend/internal/api/handlers"
	"github.com/fasim/backend/internal/api/routes"
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

var (
	port string
)

func init() {
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Fasim server",
	Long:  `Start the Factory Automation Simulator server to handle API requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func startServer() {
	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Middleware configuration
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 60 * time.Second,
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Initialize database
	database, err := db.New("fasim.db")
	if err != nil {
		e.Logger.Fatal("Failed to connect to database: ", err)
	}

	// Initialize repositories
	itemRepo := sqlite.NewItemRepository(database)
	facilityRepo := sqlite.NewFacilityRepository(database)

	// Initialize handlers
	itemHandler := handlers.NewItemHandler(itemRepo)
	facilityHandler := handlers.NewFacilityHandler(facilityRepo, itemRepo)

	// Route configuration
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Welcome to Factory Automation Simulator API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Register routes
	routes.RegisterItemRoutes(e, itemHandler)
	routes.RegisterFacilityRoutes(e, facilityHandler)

	// Start server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: e,
	}

	// Configure graceful shutdown
	go func() {
		if err := e.StartServer(server); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Failed to start server: ", err)
		}
	}()

	// Signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal("Failed to shutdown server: ", err)
	}

	log.Println("Server shutdown successfully")
}
