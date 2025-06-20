package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
	"github.com/oooJordan/WasaText/service/database"
)

// ------------------------------ #LOGIN UTENTE# ------------------------------
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

	// controllo lunghezza
	if !IsValidNickname(req.User) {
		http.Error(w, "Username must be between 3 and 16 characters", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	imageURL := scheme + "://" + host + "/defaultimage/defaultimage.jpg"

	// prendo l'id dell'utente dal database
	userID, err := rt.db.GetIdUser(req.User, imageURL)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// assegno userId
	response := LoginResponse{
		User_ID:      userID,
		ProfileImage: imageURL,
	}

	ctx.Logger.Infof("login successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// ------------------------- #AGGIORNO IMMAGINE DEL PROFILO UTENTE# --------------------------
func (rt *_router) updateProfileImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		// La risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// Token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var NewImage UpdateProfileImageRequest
	err = json.NewDecoder(r.Body).Decode(&NewImage)
	if err != nil {
		ctx.Logger.WithError(err).Error("error decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// aggiorno l'immagine di profilo nel database
	err = rt.db.UpdateProfileImage(author, NewImage.Image)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Failed to update profile image", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}

// ------------------------ #OTTENGO L'IMMAGINE DEL PROFILO DELL'UTENTE# --------------------
func (rt *_router) getProfileImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// recupero l'immagine dal database
	imageURL, err := rt.db.GetProfileImage(author)
	if err != nil {
		ctx.Logger.Error(err)
		http.Error(w, "Failed to fetch profile image", http.StatusInternalServerError)
		return
	}

	// rispondo con l'immagine
	response := map[string]string{
		"actualImage": imageURL,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// ------------------------ #OTTENGO LA LISTA DEGLI UTENTI# ---------------------------
func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	username := r.URL.Query().Get("name")

	isValid, _, err := rt.IsValidToken(r, w)
	if err != nil {
		// la risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := rt.db.GetListUsers(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error executing query")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	var convertedUsers []User
	for _, dbUser := range users { // Itero sugli utenti presi dal DB
		convertedUsers = append(convertedUsers, User{
			Nickname:     dbUser.Name,
			UserID:       dbUser.UserID,
			ProfileImage: dbUser.ProfileImage,
		})
	}
	response := struct {
		Users []User `json:"users"`
	}{
		Users: convertedUsers,
	}

	// codifico la lista degli utenti in JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Error encoding response")
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

// ----------------------- #AGGIORNO USERNAME UTENTE# -----------------------------
func (rt *_router) UpdateUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Controllo se il token è valido
	isValid, author, err := rt.IsValidToken(r, w)
	if err != nil {
		// La risposta HTTP è già gestita all'interno di IsValidToken
		return
	}
	if !isValid {
		// Token non valido
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var newUsername UpdateUsernameRequest
	err = json.NewDecoder(r.Body).Decode(&newUsername)
	if err != nil {
		ctx.Logger.WithError(err).Error("error decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !IsValidNickname(newUsername.NewUsername) {
		http.Error(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	// Aggiorno l'username nel database
	err = rt.db.UpdateUsername(author, newUsername.NewUsername)
	if err != nil {
		ctx.Logger.Error(err)
		// Se l'errore è dovuto al conflitto di username, restituisco errore 409
		if errors.Is(err, database.ErrUsernameAlreadyInUse) {
			http.Error(w, "Username already in use", http.StatusConflict) // 409
			return
		}
		// 500
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Successo
	w.WriteHeader(http.StatusNoContent)
}
