package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"servicio-gateway/client"
	"servicio-gateway/config"
)

func RegisterProfileRoutes(r *mux.Router) {
	r.HandleFunc("/profiles/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		cfg := config.LoadConfigFromEnv()
		target := cfg.ProfileURL + "/api/v1/profiles/" + id

		status, body, headers, err := client.ProxyRequest("GET", target, nil, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		CopyHeaders(w.Header(), headers)
		w.WriteHeader(status)
		w.Write(body)
	}).Methods("GET")

	r.HandleFunc("/profiles/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		cfg := config.LoadConfigFromEnv()
		target := cfg.ProfileURL + "/api/v1/profiles/" + id

		status, body, headers, err := client.ProxyRequest("PUT", target, r.Body, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		CopyHeaders(w.Header(), headers)
		w.WriteHeader(status)
		w.Write(body)
	}).Methods("PUT")
}
