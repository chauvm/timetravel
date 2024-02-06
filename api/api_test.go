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

func makeRequestV2(req *http.Request) *httptest.ResponseRecorder {
	// sql test db
	db, err := database.CreateConnection()
	if err != nil {
		panic(err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	// v2
	persistentService := service.NewPersistentRecordService(db)
	apiV2 := NewAPIV2(&persistentService)
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiV2.CreateRoutes(apiRouteV2)

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
	req, _ := http.NewRequest("GET", "/api/v2/records/1", nil)
	rr := makeRequestV2(req)
	assert.Equal(t, 400, rr.Code)
	assert.Equal(t, `{"id":1,"data":{}}`, rr.Body.String())
}

func TestPostRecordsV2(t *testing.T) {
	var jsonStr = []byte(`{"hello":"world"}`)
	req, _ := http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer(jsonStr))
	rr := makeRequestV2(req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}\n", rr.Body.String())
	// ^ FAIL, currently had extra ,\"accumulated\":null,\"version\":0,\"timestamp\":\"\"}
}

func TestGetVersions(t *testing.T) {
	// create a couple of versions of a record
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))

	req, _ := http.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rr := makeRequestV2(req)
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "{\"data\":[1, 2, 3]}", rr.Body.String())
}

func TestGetRecordAtTimestamp(t *testing.T) {
	// create a couple of versions of a record
	// TODO: freeze time
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world"}`)))
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":"world 2","status":"ok"}`)))
	http.NewRequest("POST", "/api/v2/records/1", bytes.NewBuffer([]byte(`{"hello":null}`)))

	// assert data at each version
	req1, _ := http.NewRequest("GET", "/api/v2/records/1/1707222000", nil)
	rr1 := makeRequestV2(req1)
	assert.Equal(t, 200, rr1.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"hello\":\"world\"}\n", rr1.Body.String())

	req2, _ := http.NewRequest("GET", "/api/v2/records/1/1707222001", nil)
	rr2 := makeRequestV2(req2)
	assert.Equal(t, 200, rr2.Code)
	assert.Equal(t, "{\"id\":1,\"data\":\"hello\":\"world 2\",\"status\":\"ok\"}\n", rr2.Body.String())

	req3, _ := http.NewRequest("GET", "/api/v2/records/1/1707222002", nil)
	rr3 := makeRequestV2(req3)
	assert.Equal(t, 200, rr3.Code)
	assert.Equal(t, "{\"id\":1,\"data\":{\"status\":\"ok\"}\n", rr3.Body.String())
}
