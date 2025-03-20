package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
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
