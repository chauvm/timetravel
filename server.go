package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chauvm/timetravel/api"
	"github.com/chauvm/timetravel/database"
	"github.com/chauvm/timetravel/service"
	"github.com/gorilla/mux"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	log.Println("main: starting server...")
	router := mux.NewRouter()

	// v1: I realized I misunderstood the requirements
	// I keep v1 data in memory, should have been also
	// persistent in SQLite
	imMemoryService := service.NewInMemoryRecordService()
	apiV1 := api.NewAPI(&imMemoryService)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	apiV1.CreateRoutes(apiRoute)

	// v2
	log.Println("main: create database connection")
	db, err := database.CreateConnection()
	log.Println("main: database connection created")

	if err != nil {
		log.Fatal(err)
	}

	persistentService := service.NewPersistentRecordService(db)
	apiV2 := api.NewAPIV2(&persistentService)
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiRouteV2.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	apiV2.CreateRoutes(apiRouteV2)

	address := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("listening on %s", address)
	log.Fatal(srv.ListenAndServe())
}
