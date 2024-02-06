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

	// apiV1 := api.NewAPI(&imMemoryService)

	// apiV1.CreateRoutes(apiRoute)

	// v2
	log.Println("main: create database connection")
	db, err := database.CreateConnection()
	log.Println("main: database connection created")

	if err != nil {
		log.Fatal(err)
	}

	persistentService := service.NewPersistentRecordService(db)

	newAPI := api.NewAPI(&persistentService)
	newAPIV2 := api.NewAPIV2(&persistentService)

	apiRouteV1 := router.PathPrefix("/api/v1").Subrouter()
	apiRouteV1.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiRouteV2.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})

	newAPI.CreateRoutes(apiRouteV1)
	newAPIV2.CreateRoutes(apiRouteV2)

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
