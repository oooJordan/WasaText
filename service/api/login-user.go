package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// Gestione del login e creazione dell'utente
func (rt *_router) loginUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Decoding the JSON containing the username
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validating the username
	if req.User == "" {
		http.Error(w, "Missing 'Username' parameter", http.StatusBadRequest)
		return
	}

	// Getting the UId from the database
	userID, err := rt.db.GetIdUser(req.User)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Creating the response
	response := LoginResponse{
		User_ID: userID,
	}

	//ctx.Logger.Infof()
	ctx.Logger.Infof("login successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
