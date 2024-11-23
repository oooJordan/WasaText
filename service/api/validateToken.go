package api

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("mySuperSecretKey") // Chiave segreta per firmare il JWT

// Genera un token JWT per un utente
func GenerateJWT(userId int) (string, error) {
	claims := jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(24 * time.Hour).Unix(), // Scadenza: 24 ore
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Verifica un token JWT e restituisce l'userId
func VerifyJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo di firma non valido")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("token non valido: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["userId"].(float64)), nil
	}

	return 0, fmt.Errorf("token non valido o scaduto")
}
