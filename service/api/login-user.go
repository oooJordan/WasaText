package api

import (
	//"database/sql"
	"encoding/json"
	//"errors"
	"net/http"
	//"os"
	//"path/filepath"
	//"strconv"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) loginUser(w http.ResponseWriter, r *http.Request, x httprouter.Params) {
	w.Header().Set("content-type", "text/plain")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
