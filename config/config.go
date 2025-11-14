package config

import (
	"log"
	"os"
)

// Config expuesto para que handlers lo usen (campos exportados)
type Config struct {
	SecurityURL string
	ProfileURL  string
	EventBusURL string
	Port        string
}

// LoadConfigFromEnv carga variables de entorno y devuelve Config
func LoadConfigFromEnv() Config {
	cfg := Config{
		SecurityURL: os.Getenv("SECURITY_URL"),
		ProfileURL:  os.Getenv("PROFILE_URL"),
		EventBusURL: os.Getenv("EVENT_BUS_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8088"
	}

	// Logging útil para debugging
	if cfg.SecurityURL == "" {
		log.Println("[WARN] SECURITY_URL no configurada")
	}
	if cfg.ProfileURL == "" {
		log.Println("[WARN] PROFILE_URL no configurada")
	}
	if cfg.EventBusURL == "" {
		log.Println("[WARN] EVENT_BUS_URL no configurada (eventos deshabilitados)")
	}

	return cfg
}
