package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/chauvm/timetravel/entity"
	"github.com/chauvm/timetravel/service"
	"github.com/gorilla/mux"
)

// v1 POST /records/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
func (a *API) PostRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	// first retrieve the record
	record, err := a.records.GetRecord(
		ctx,
		int(idNumber),
	)

	if !errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		record, err = a.records.UpdateRecord(ctx, int(idNumber), body)
	} else { // record does not exist

		// exclude the delete updates
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}

		record = entity.Record{
			ID:   int(idNumber),
			Data: recordMap,
		}
		err = a.records.CreateRecord(ctx, record)
	}

	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}

// v2 POST /records/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
func (a *APIV2) PostRecords(w http.ResponseWriter, r *http.Request) {
	log.Print("PostRecords v2")
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	// first retrieve the record
	record, err := a.records.GetRecord(
		ctx,
		int(idNumber),
	)
	log.Printf("PostRecords v2: record: %v", record)

	if !errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		record, err = a.records.UpdateRecord(ctx, int(idNumber), body)
	} else { // record does not exist
		log.Print("PostRecords v2: record does not exist")

		// exclude the delete updates
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}

		record = entity.Record{
			ID:   int(idNumber),
			Data: recordMap,
			// accumulated data is the same as the data in a new record
			Accumulated: recordMap,
			Version:     1,
		}
		err = a.records.CreateRecord(ctx, record)
	}

	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
