package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const VERSION = "1.0.0"

var startTime = time.Now()

// ------------------------------
// Health Response Types
// ------------------------------
type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	Uptime        string `json:"uptime"`
	UptimeSeconds int64  `json:"uptimeSeconds"`
}

type HealthCheck struct {
	Data   map[string]interface{} `json:"data"`
	Name   string                 `json:"name"`
	Status string                 `json:"status"`
}

type HealthResponseWithChecks struct {
	Status        string        `json:"status"`
	Checks        []HealthCheck `json:"checks"`
	Version       string        `json:"version"`
	Uptime        string        `json:"uptime"`
	UptimeSeconds int64         `json:"uptimeSeconds"`
}

// ------------------------------
// Helper para formatear uptime
// ------------------------------
func formatUptime(duration time.Duration) string {
	seconds := int64(duration.Seconds())
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// ------------------------------
// Handlers de Health
// ------------------------------
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	checks := []HealthCheck{
		{
			Data: map[string]interface{}{
				"from":   startTime.Format(time.RFC3339Nano),
				"status": "READY",
			},
			Name:   "Readiness check",
			Status: "UP",
		},
		{
			Data: map[string]interface{}{
				"from":   startTime.Format(time.RFC3339Nano),
				"status": "ALIVE",
			},
			Name:   "Liveness check",
			Status: "UP",
		},
	}

	response := HealthResponseWithChecks{
		Status:        "UP",
		Checks:        checks,
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:        "READY",
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	response := HealthResponse{
		Status:        "LIVE",
		Version:       VERSION,
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ------------------------------
// Main
// ------------------------------
func main() {
	// Configuración desde variables de entorno
	cfg := loadConfigFromEnv()

	// Cliente HTTP global (definido en client.go)
	httpClient = &http.Client{Timeout: 10 * time.Second}

	r := mux.NewRouter()

	// ------------------ ROUTES ---------------------

	// ==== Rutas de seguridad ====
	r.HandleFunc("/auth/login", makeProxyToSecurity("POST", "/auth/login")).Methods("POST")
	r.HandleFunc("/auth/register", makeProxyToSecurity("POST", "/auth/register")).Methods("POST")

	// ==== Operaciones CRUD Usuario ====
	r.HandleFunc("/users/{id}", handleDeleteUser).Methods("DELETE")
	r.HandleFunc("/users/{id}", handleGetUserFull).Methods("GET")
	r.HandleFunc("/users/{id}", handleUpdateUserFull).Methods("PUT")

	// ==== Health ====
	r.HandleFunc("/health", HealthHandler).Methods("GET")
	r.HandleFunc("/ready", ReadyHandler).Methods("GET")
	r.HandleFunc("/live", LiveHandler).Methods("GET")

	// ------------------------------------------------

	addr := ":" + cfg.Port
	log.Printf("servicio-gateway escuchando en %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
