package main

import (
	"net/http"
	"os"

	"delivery/database"
	_ "delivery/database/migrations" // Import migrations package for init() registration
	"delivery/handler"
	"delivery/helper"
	"delivery/services/queue"

	"github.com/gorilla/mux"
)

func main() {
	helper.InitLogger()

	// Connect to databases
	dbConn, err := database.Connect()
	if err != nil {
		helper.Log.Fatalf("Failed to connect to databases: %v", err)
		return
	}

	// Access writer and reader databases
	db := dbConn.Writer
	readerDB := dbConn.Reader

	// Apply database migrations
	err = database.ApplyDatabaseUpdates(db, "./database/migrations")
	if err != nil {
		helper.Log.Fatalf("Failed to apply database migrations: %v", err)
		return
	}

	// Create and configure router
	r := mux.NewRouter()

	// Register health check route
	handler.RegisterHealthRoute(r)

	// Initialize and start consumers
	consumerManager, err := queue.NewConsumerManager(db, readerDB)
	if err != nil {
		helper.Log.Fatalf("Failed to initialize consumer manager: %v", err)
	}
	defer consumerManager.Close()

	err = consumerManager.StartConsumers()
	if err != nil {
		helper.Log.Fatalf("Failed to start consumers: %v", err)
	}

	// Register API routes for WhatsApp, Email, SMS, Providers, and Templates
	handler.RegisterWhatsAppRoutes(r, db, readerDB, consumerManager.GetPulsarClient())
	handler.RegisterEmailRoutes(r, db)
	handler.RegisterSMSRoutes(r, db, readerDB, consumerManager.GetPulsarClient())
	handler.RegisterProviderRoutes(r, db, readerDB)
	handler.RegisterTemplateRoutes(r, db, readerDB)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	helper.Log.Infof("Starting server on port %s...", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		helper.Log.Fatalf("Failed to start server: %v", err)
	}
}
