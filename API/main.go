package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jonathanface/storm-reporter/API/dao"
	"github.com/jonathanface/storm-reporter/API/middleware"
	"github.com/jonathanface/storm-reporter/API/routes"
)

func main() {
	// Validate required environment variables
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB")
	mongoColl := os.Getenv("MONGO_COLL")

	if kafkaBrokers == "" || mongoURI == "" || mongoDBName == "" || mongoColl == "" {
		log.Fatal("Environment variables KAFKA_BROKERS, MONGO_URI, MONGO_DB, and MONGO_COLL must be set")
	}

	// Initialize DAO
	dao, err := dao.NewStormDAO(mongoURI, mongoDBName, mongoColl)
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}
	defer dao.Disconnect()

	// routes with middleware
	mux := http.NewServeMux()
	middlewareContext := middleware.WithDAOContext(dao)
	mux.Handle("/messages", middlewareContext(routes.GetMessagesHandler))

	// Start
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
