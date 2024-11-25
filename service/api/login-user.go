package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

// Gestione del login e creazione dell'utente
func (rt *_router) loginUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// controllo username
	if req.User == "" {
		http.Error(w, "Missing 'Username' parameter", http.StatusBadRequest)
		return
	}

	// prendo l'id dell'utente dal database
	userID, err := rt.db.GetIdUser(req.User)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// assegno userId
	response := LoginResponse{
		User_ID: userID,
	}

	ctx.Logger.Infof("login successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
