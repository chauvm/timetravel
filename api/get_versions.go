package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// v2 GET /records/{id}/versions
// GetVersions lists all versions of the record.
func (a *APIV2) GetVersions(w http.ResponseWriter, r *http.Request) {
	// TODO
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	versions, err := a.records.GetRecordVersions(
		ctx,
		int(idNumber),
	)
	if err != nil {
		err := writeError(w, fmt.Sprintf("Unable to retrieve versions for record %d", idNumber), http.StatusBadRequest)
		logError(err)
		return
	}

	response := map[string]interface{}{"data": versions}

	err = writeJSON(w, response, http.StatusOK)
	logError(err)
}
