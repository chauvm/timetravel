package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// v2 GET /records/{record_id}/{version}
// Get record at a specific version
func (a *APIV2) GetRecordAtVersion(w http.ResponseWriter, r *http.Request) {
	// TODO
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	version := mux.Vars(r)["version"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	versionNumber, err := strconv.ParseInt(version, 10, 32)

	if err != nil || versionNumber <= 0 {
		err := writeError(w, "invalid version; version must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	record, err := a.records.GetRecordAtVersion(
		ctx,
		int(idNumber),
		int(versionNumber),
	)
	if err != nil {
		err := writeError(w, fmt.Sprintf("Unable to retrieve record of id %d at version %d", idNumber, versionNumber), http.StatusBadRequest)
		logError(err)
		return
	}

	returnedRecord := record.GetExternalRecord()

	err = writeJSON(w, returnedRecord, http.StatusOK)
	logError(err)
}
