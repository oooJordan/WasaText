package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
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

// var emojiCodeRegex = regexp.MustCompile(`^[A-Za-z0-9+]+$`)

func IsValidNickname(nick string) bool {
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}

	return validNickname.MatchString(nick)
}

/*
	func isValidMessage(mex string) bool {
		if len(mex) < 1 || len(mex) > 200 {
			return false
		}
		return true
	}

	func isValidEmoji(emojicode string) bool {
		if len(emojicode) < 1 || len(emojicode) > 20 {
			return false
		}
		return emojiCodeRegex.MatchString(emojicode)
	}
*/
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

func (rt *_router) uploadImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, "Failed to parse form. Ensure the file is below 10 MB.", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	if len(fileBytes) > 10*1024*1024 {
		http.Error(w, "File too large. Maximum allowed size is 10 MB.", http.StatusRequestEntityTooLarge)
		return
	}

	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" && fileType != "image/png" {
		http.Error(w, "Invalid file type. Only JPEG and PNG are supported.", http.StatusUnsupportedMediaType)
		return
	}

	savedFilePath, err := saveFile(fileBytes, header)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"imageUrl": savedFilePath,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode response")
	}
}

func saveFile(fileBytes []byte, header *multipart.FileHeader) (string, error) {
	dir := "foto"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return "", err
		}
	}

	fileName := time.Now().Format("20060102150405") + "-" + header.Filename
	filePath := filepath.Join(dir, fileName)

	err := os.WriteFile(filePath, fileBytes, 0644)
	if err != nil {
		return "", err
	}

	return "/foto/" + fileName, nil
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
