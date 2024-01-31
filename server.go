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
	router := mux.NewRouter()

	// v1
	service := service.NewInMemoryRecordService()
	api := api.NewAPI(&service)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	api.CreateRoutes(apiRoute)

	// v2
	db := database.connection.CreateConnection()
	apiV2 := api.NewAPIV2(&db)
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiRouteV2.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})

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
