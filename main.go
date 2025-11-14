package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// ...existing code...
	// Cargar configuración mínima desde variables de entorno
	cfg := loadConfigFromEnv()

	// Cliente HTTP compartido con timeout
	httpClient = &http.Client{Timeout: 10 * time.Second}

	r := mux.NewRouter()

	// Autenticación y registro (proxy directo a security)
	r.HandleFunc("/auth/login", makeProxyToSecurity("POST", "/auth/login")).Methods("POST")
	r.HandleFunc("/auth/register", makeProxyToSecurity("POST", "/auth/register")).Methods("POST")

	// Eliminación de usuario: reenvía a security y genera evento
	r.HandleFunc("/users/{id}", handleDeleteUser).Methods("DELETE")

	// Consulta y actualización de datos completos del usuario
	r.HandleFunc("/users/{id}", handleGetUserFull).Methods("GET")
	r.HandleFunc("/users/{id}", handleUpdateUserFull).Methods("PUT")

	addr := ":" + cfg.Port
	log.Printf("servicio-gateway escuchando en %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
