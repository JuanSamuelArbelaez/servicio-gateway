package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

var startTime = time.Now()

const VERSION = "1.0.0"

func Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	resp := map[string]interface{}{
		"status":        "UP",
		"version":       VERSION,
		"uptime":        uptime.String(),
		"uptimeSeconds": int64(uptime.Seconds()),
	}

	json.NewEncoder(w).Encode(resp)
}
