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

func makeRequestV1(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	// v1
	imMemoryService := service.NewInMemoryRecordService()
	apiV1 := NewAPI(&imMemoryService)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiV1.CreateRoutes(apiRoute)

	router.ServeHTTP(rr, req)

	return rr

	// if body := response.Body.String(); body != "[]" {
	// 	t.Errorf("Expected an empty array. Got %s", body)
	// }
}

func setUpV2() *mux.Router {
	// sql test db
	db, err := database.CreateConnectionUnitTests()
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	// v2
	persistentService := service.NewPersistentRecordService(db)
	apiV2 := NewAPIV2(&persistentService)
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiV2.CreateRoutes(apiRouteV2)

	return router
}

func makeRequestV2(router *mux.Router, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr

	// if body := response.Body.String(); body != "[]" {
	// 	t.Errorf("Expected an empty array. Got %s", body)
	// }
}

// GET /api/v1/records/{id}
func TestGetRecordsV1(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/records/1", nil)
	rr := makeRequestV1(req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"record of id 1 does not exist\"}\n", rr.Body.String())
}

// TODO 2: fix POST v1 to return as is
// POST /api/v1/records/{id}
func TestPostRecordsV1(t *testing.T) {
	var jsonStr = []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/api/v1/records/1", bytes.NewBuffer(jsonStr))
	rr := makeRequestV1(req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}\n", rr.Body.String())
	// ^ FAIL, currently had extra ,\"accumulated\":null,\"version\":0,\"timestamp\":\"\"}
}

// TODO 1: fix GET
// GET /api/v2/records/{id}
func TestGetRecordsV2(t *testing.T) {
	router := setUpV2()
	// get a record not yet exist
	req, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr := makeRequestV2(router, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"record of id 1 does not exist\"}\n", rr.Body.String())
}

func TestPostRecordsV2(t *testing.T) {
	router := setUpV2()
	// create a new record
	var jsonStr = []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr))
	rr := makeRequestV2(router, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr.Body.String())

	// confirm can GET the record
	req1, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr1 := makeRequestV2(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr1.Body.String())

	// update the record
	var jsonStr2 = []byte(`{"hello":"world 2","status":"ok"}`)
	req2, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr2))
	rr2 := makeRequestV2(router, req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr2.Body.String())

	// confirm the record is updated
	req3, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr3 := makeRequestV2(router, req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr3.Body.String())

	// update the record with null
	var jsonStr3 = []byte(`{"hello":null}`)
	req4, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr3))
	rr4 := makeRequestV2(router, req4)
	assert.Equal(t, 200, rr4.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr4.Body.String())

	// confirm the record is updated
	req5, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr5 := makeRequestV2(router, req5)
	assert.Equal(t, 200, rr5.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr5.Body.String())
}

func TestGetVersions(t *testing.T) {
	router := setUpV2()
	// a non-existing record should return an empty list without throwing an error
	req, _ := http.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rr := makeRequestV2(router, req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"data\":[]}\n", rr.Body.String())

	// create a couple of versions of a record
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	makeRequestV2(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	makeRequestV2(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))
	makeRequestV2(router, req)

	req1, _ := http.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rr1 := makeRequestV2(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"data\":[3,2,1]}\n", rr1.Body.String())
}

func TestGetRecordAtVersion(t *testing.T) {
	router := setUpV2()
	// a non-existing record should return an error
	req, _ := http.NewRequest("GET", "/api/v2/records/1/1", nil)
	rr := makeRequestV2(router, req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, "{\"error\":\"Unable to retrieve record of id 1 at version 1\"}\n", rr.Body.String())

	// create a couple of versions of a record
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	makeRequestV2(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	makeRequestV2(router, req)
	req, _ = http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))
	makeRequestV2(router, req)

	// assert data at each version
	req1, _ := http.NewRequest("GET", "/api/v2/records/1/1", nil)
	rr1 := makeRequestV2(router, req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}}\n", rr1.Body.String())

	req2, _ := http.NewRequest("GET", "/api/v2/records/1/2", nil)
	rr2 := makeRequestV2(router, req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world 2\",\"status\":\"ok\"}}\n", rr2.Body.String())

	req3, _ := http.NewRequest("GET", "/api/v2/records/1/3", nil)
	rr3 := makeRequestV2(router, req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}}\n", rr3.Body.String())
}
