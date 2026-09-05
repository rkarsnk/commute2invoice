// Package main は、アプリケーションのエントリーポイントです。
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/config"
	"github.com/rkarsnk/commute2invoice/src/db"
	"github.com/rkarsnk/commute2invoice/src/handler"
	"github.com/rkarsnk/commute2invoice/src/middleware"
	"github.com/rkarsnk/commute2invoice/src/repository"
	"github.com/rkarsnk/commute2invoice/src/service"
)

func main() {
	cfg := config.Load()
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close(database)

	repo := repository.NewSQLiteRepository(database)
	fareService := service.NewFareService(repo)
	tripService := service.NewTripService(repo, repo)
	invoiceService := service.NewInvoiceService(repo, repo, repo)

	router := setupRouter(
		handler.NewFareHandler(fareService),
		handler.NewTripHandler(tripService),
		handler.NewInvoiceHandler(invoiceService),
		handler.NewCalendarHandler(tripService, invoiceService),
	)

	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(
	fareHandler *handler.FareHandler,
	tripHandler *handler.TripHandler,
	invoiceHandler *handler.InvoiceHandler,
	calendarHandler *handler.CalendarHandler,
) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.NoCacheMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())

	calendarGroup := router.Group("/calendar")
	calendarGroup.GET("", calendarHandler.GetCalendar)
	calendarGroup.GET("/:date/trips", tripHandler.ListTripsByDate)
	calendarGroup.POST("/:date/trips", tripHandler.CreateTrip)

	tripGroup := router.Group("/trips")
	tripGroup.PUT("/:id", tripHandler.UpdateTrip)
	tripGroup.DELETE("/:id", tripHandler.DeleteTrip)

	fareEvidenceGroup := router.Group("/setup/fare-evidence")
	fareEvidenceGroup.POST("", fareHandler.CreateFareEvidence)
	fareEvidenceGroup.GET("", fareHandler.ListValidFareEvidences)
	fareEvidenceGroup.GET("/:id/image", fareHandler.GetFareEvidenceImage)
	router.GET("/fare-evidences", fareHandler.ListAllFareEvidences)
	router.PATCH("/fare-evidences/:id/validity", fareHandler.UpdateFareEvidenceValidity)
	router.DELETE("/fare-evidences/:id", fareHandler.DeleteFareEvidence)

	invoiceGroup := router.Group("/invoices")
	invoiceGroup.GET("/preview", invoiceHandler.GetInvoicePreview)
	invoiceGroup.POST("/confirm", invoiceHandler.ConfirmInvoice)
	invoiceGroup.GET("/:year/:month/:billing_target/pdf", invoiceHandler.GetInvoicePDF)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Static("/public", "./src/frontend")
	router.StaticFile("/", "./src/frontend/index.html")
	router.NoRoute(middleware.NotFoundMiddleware())

	return router
}
