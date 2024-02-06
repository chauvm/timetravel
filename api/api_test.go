package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chauvm/timetravel/database"
	"github.com/chauvm/timetravel/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setUp() *mux.Router {
	// sql test db
	db, err := database.CreateConnectionUnitTests()
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	// v2
	persistentService := service.NewPersistentRecordService(db)

	newAPI := NewAPI(&persistentService)
	newAPIV2 := NewAPIV2(&persistentService)

	apiRouteV1 := router.PathPrefix("/api/v1").Subrouter()
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()

	newAPI.CreateRoutes(apiRouteV1)
	newAPIV2.CreateRoutes(apiRouteV2)

	return router
}

func makeRequest(router *mux.Router, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

// GET /api/v1/records/{id}
func TestGetRecordsV1(t *testing.T) {
	router := setUp()
	req, _ := http.NewRequest("GET", "/api/v1/records/1", nil)
	rr := makeRequest(router, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"record of id 1 does not exist\"}\n", rr.Body.String())
}

// TODO 2: fix POST v1 to return as is
// POST /api/v1/records/{id}
func TestPostRecordsV1(t *testing.T) {
	router := setUp()

	var jsonStr = []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/api/v1/records/1", bytes.NewBuffer(jsonStr))
	rr := makeRequest(router, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr.Body.String())

	// confirm can GET the record
	req1, _ := http.NewRequest("GET", "/api/v1/records/1", nil)
	rr1 := makeRequest(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr1.Body.String())

	// update the record
	var jsonStr2 = []byte(`{"hello":"world 2","status":"ok"}`)
	req2, _ := http.NewRequest("POST", "/api/v1/records/1", bytes.NewBuffer(jsonStr2))
	rr2 := makeRequest(router, req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr2.Body.String())

	// confirm the record is updated
	req3, _ := http.NewRequest("GET", "/api/v1/records/1", nil)
	rr3 := makeRequest(router, req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr3.Body.String())

	// update the record with null
	var jsonStr3 = []byte(`{"hello":null}`)
	req4, _ := http.NewRequest("POST", "/api/v1/records/1", bytes.NewBuffer(jsonStr3))
	rr4 := makeRequest(router, req4)
	assert.Equal(t, 200, rr4.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr4.Body.String())

	// confirm the record is updated
	req5, _ := http.NewRequest("GET", "/api/v1/records/1", nil)
	rr5 := makeRequest(router, req5)
	assert.Equal(t, 200, rr5.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr5.Body.String())
}

// TODO 1: fix GET
// GET /api/v2/records/{id}
func TestGetRecordsV2(t *testing.T) {
	router := setUp()
	// get a record not yet exist
	req, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr := makeRequest(router, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"record of id 1 does not exist\"}\n", rr.Body.String())
}

func TestPostRecordsV2(t *testing.T) {
	router := setUp()
	// create a new record
	var jsonStr = []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr))
	rr := makeRequest(router, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr.Body.String())

	// confirm can GET the record
	req1, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr1 := makeRequest(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr1.Body.String())

	// update the record
	var jsonStr2 = []byte(`{"hello":"world 2","status":"ok"}`)
	req2, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr2))
	rr2 := makeRequest(router, req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr2.Body.String())

	// confirm the record is updated
	req3, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr3 := makeRequest(router, req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr3.Body.String())

	// update the record with null
	var jsonStr3 = []byte(`{"hello":null}`)
	req4, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr3))
	rr4 := makeRequest(router, req4)
	assert.Equal(t, 200, rr4.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr4.Body.String())

	// confirm the record is updated
	req5, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr5 := makeRequest(router, req5)
	assert.Equal(t, 200, rr5.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr5.Body.String())
}

func TestGetVersions(t *testing.T) {
	router := setUp()
	// a non-existing record should return an empty list without throwing an error
	req, _ := http.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rr := makeRequest(router, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"data\":[]}\n", rr.Body.String())

	// create a couple of versions of a record
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	makeRequest(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	makeRequest(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))
	makeRequest(router, req)

	req1, _ := http.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rr1 := makeRequest(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"data\":[3,2,1]}\n", rr1.Body.String())
}

func TestGetRecordAtVersion(t *testing.T) {
	router := setUp()
	// a non-existing record should return an error
	req, _ := http.NewRequest("GET", "/api/v2/records/1/1", nil)
	rr := makeRequest(router, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"Unable to retrieve record of id 1 at version 1\"}\n", rr.Body.String())

	// create a couple of versions of a record
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	makeRequest(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	makeRequest(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))
	makeRequest(router, req)

	// assert data at each version
	req1, _ := http.NewRequest("GET", "/api/v2/records/1/1", nil)
	rr1 := makeRequest(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr1.Body.String())

	req2, _ := http.NewRequest("GET", "/api/v2/records/1/2", nil)
	rr2 := makeRequest(router, req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr2.Body.String())

	req3, _ := http.NewRequest("GET", "/api/v2/records/1/3", nil)
	rr3 := makeRequest(router, req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr3.Body.String())
}
