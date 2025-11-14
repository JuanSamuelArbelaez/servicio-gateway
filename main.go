package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"servicio-gateway/client"
	"servicio-gateway/config"
	"servicio-gateway/handlers"
)

func main() {
	cfg := config.LoadConfigFromEnv()

	// Configurar cliente http global
	client.HttpClient = &http.Client{Timeout: 10 * time.Second}

	r := mux.NewRouter()

	// CORS middleware (func CORS defined in root cors.go)
	r.Use(CORS)

	// Register public routes (auth, user CRUD proxies)
	handlers.RegisterUserServiceRoutes(r)

	// Protected subrouter (jwt)
	api := r.PathPrefix("/").Subrouter()
	api.Use(JWTMiddleware)

	// Profile routes (protected)
	handlers.RegisterProfileRoutes(api)

	// Composite endpoints (protected)
	api.HandleFunc("/users/{id}", handlers.HandleGetUserFull).Methods("GET")
	api.HandleFunc("/users/{id}", handlers.HandleUpdateUserFull).Methods("PUT")
	api.HandleFunc("/users/{id}", handlers.HandleDeleteUser).Methods("DELETE")

	// Health endpoints (public)
	r.HandleFunc("/health", handlers.Health).Methods("GET")
	r.HandleFunc("/ready", handlers.Health).Methods("GET")
	r.HandleFunc("/live", handlers.Health).Methods("GET")

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Gateway escuchando en %s", addr)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}
