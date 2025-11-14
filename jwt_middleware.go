package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"context"
	"github.com/golang-jwt/jwt/v5"
)

// Clave secreta cargada por ENV
var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[WARN] JWT_SECRET no está definido, usando valor por defecto de desarrollo")
		secret = "default-dev-secret"
	}
	jwtSecret = []byte(secret)
}

// ---------------------------------------------------------
// Extraer token desde header Authorization
// ---------------------------------------------------------
func extractToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid Authorization format (use Bearer <token>)")
	}

	return parts[1], nil
}

// ---------------------------------------------------------
// Validar token y devolver claims
// ---------------------------------------------------------
func validateJWT(tokenString string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validar método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ---------------------------------------------------------
// Middleware: validar JWT y meter claims en el contexto
// ---------------------------------------------------------
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Insertar claims en el contexto
		ctx := context.WithValue(r.Context(), "tokenData", claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------
// Helper para obtener los claims en cualquier handler
// ---------------------------------------------------------
func GetTokenData(r *http.Request) map[string]interface{} {
	claims := r.Context().Value("tokenData")
	if claims == nil {
		return nil
	}
	return claims.(jwt.MapClaims)
}

// ---------------------------------------------------------
// Utilidad por si necesitas responder JSON
// ---------------------------------------------------------
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
