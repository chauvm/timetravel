package api

import (
	"github.com/chauvm/timetravel/service"
	"github.com/gorilla/mux"
)

type API struct {
	records service.RecordService
}

func NewAPI(records service.RecordService) *API {
	return &API{records}
}

// generates all api routes
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}

type APIV2 struct {
	records service.RecordService
}

func NewAPIV2(records service.RecordService) *APIV2 {
	return &APIV2{records}
}

// generates all api routes
func (a *APIV2) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
	// new endpoints compared to v1
	routes.Path("/records/{record_id}/versions").HandlerFunc(a.GetVersions).Methods("GET")
	routes.Path("/records/{record_id}/{timestamp}").HandlerFunc(a.GetRecordAtTimestamp).Methods("GET")
}
