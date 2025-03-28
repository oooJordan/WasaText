package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/oooJordan/WasaText/service/api/reqcontext"
)

var validNickname = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func IsValidNickname(nick string) bool {
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}

	return validNickname.MatchString(nick)
}

func (rt *_router) IsValidToken(r *http.Request, w http.ResponseWriter) (bool, int, error) {
	// valore dell'intestazione Authorization della richiesta HTTP
	authHeader := r.Header.Get("Authorization")

	// controllo se non è vuoto
	if len(authHeader) == 0 {
		http.Error(w, "Authorization Header missing", http.StatusForbidden)
		return false, 0, errors.New("authorization Header missing")
	}
	// divido la stringa in base agli spazi es:["Barer", "user"] e cnontrollo se c'è barer
	HeaderPart := strings.Fields(authHeader)
	if len(HeaderPart) < 2 || HeaderPart[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusForbidden)
		return false, 0, errors.New("invalid Authorization format")
	}
	token := HeaderPart[1] // prendo il token

	userid := ExtractUserIdFromToken(token) // id_utente

	found, _, err := rt.db.CheckIDDatabase(userid)
	if err != nil || !found {
		http.Error(w, "Non autorizzato", http.StatusUnauthorized)
		return false, 0, nil
	}
	return true, userid, nil
}

func ExtractUserIdFromToken(token string) int {
	id, _ := strconv.Atoi(token)
	return id
}

// ------------#CONTROLLO SE L'UTENTE È NELLA CONVERSAZIONE#----------------
func (rt *_router) isUserInConversation(conversationID int, userID int, chatType string) (bool, error) {
	switch chatType {
	case "group_chat":
		return rt.db.IsUserInGroup(conversationID, userID)
	case "private_chat":
		return rt.db.IsUserInPrivateChat(conversationID, userID)
	default:
		return false, errors.New("invalid conversation type")
	}
}

func getIntParam(w http.ResponseWriter, ps httprouter.Params, paramName string) (int, bool) {
	paramStr := ps.ByName(paramName)
	param, err := strconv.Atoi(paramStr)
	if err != nil {
		http.Error(w, "Invalid "+paramName, http.StatusBadRequest)
		return 0, false
	}
	return param, true
}

func (rt *_router) uploadImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	isValid, _, err := rt.IsValidToken(r, w)
	if err != nil {
		return
	}
	if !isValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	const maxUploadSize = 10 << 20 // 10MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadSize))

	if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
		http.Error(w, "upload too large", http.StatusRequestEntityTooLarge)
		return
	}

	uploadedFile, fileInfo, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	magicBytes := make([]byte, 512)
	if _, err := uploadedFile.Read(magicBytes); err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}

	fileType := http.DetectContentType(magicBytes)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowedTypes[fileType] {
		http.Error(w, "invalid image format", http.StatusUnsupportedMediaType)
		return
	}

	if _, err := uploadedFile.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to seek file", http.StatusInternalServerError)
		return
	}

	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "failed to create upload directory", http.StatusInternalServerError)
		return
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	finalName := timestamp + "_" + fileInfo.Filename
	finalPath := filepath.Join(uploadDir, finalName)

	dstFile, err := os.Create(finalPath)
	if err != nil {
		http.Error(w, "cannot store the image", http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, uploadedFile); err != nil {
		http.Error(w, "file saving failed", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	imageURL := scheme + "://" + host + "/uploads/" + finalName

	response := struct {
		ImageURL string `json:"imageUrl"`
	}{
		ImageURL: imageURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
