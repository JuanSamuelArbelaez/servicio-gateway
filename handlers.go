package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// -------------------------------------------
// PROXY BÁSICO
// -------------------------------------------
func makeProxyToSecurity(method, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := loadConfigFromEnv()
		target := cfg.SecurityURL + path
		status, body, headers, err := proxyRequest(method, target, r.Body, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		copyHeaders(w.Header(), headers)
		w.WriteHeader(status)
		w.Write(body)
	}
}

// -------------------------------------------
// DELETE USER (+ evento user.deleted)
// -------------------------------------------
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromEnv()
	vars := mux.Vars(r)
	id := vars["id"]
	target := cfg.SecurityURL + "/users/" + id

	status, body, headers, err := proxyRequest("DELETE", target, nil, r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Emitir evento SOLO si security eliminó correctamente
	if status >= 200 && status < 300 {
		event := map[string]interface{}{
			"type": "user.deleted",
			"payload": map[string]string{
				"userId": id,
			},
		}

		if err := postEvent(cfg, event); err != nil {
			log.Printf("falló publicar evento de eliminación: %v", err)
		}
	}

	copyHeaders(w.Header(), headers)
	w.WriteHeader(status)
	w.Write(body)
}

// -------------------------------------------
// GET USER — UNIFICAR SECURITY + PROFILE
// -------------------------------------------
func handleGetUserFull(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromEnv()
	vars := mux.Vars(r)
	id := vars["id"]

	// Llamada a security
	secURL := cfg.SecurityURL + "/users/" + id
	statusS, bodyS, _, errS := proxyRequest("GET", secURL, nil, r.Header)
	if errS != nil {
		http.Error(w, errS.Error(), http.StatusBadGateway)
		return
	}

	// Llamada a profile
	profURL := cfg.ProfileURL + "/profiles/" + id
	statusP, bodyP, _, errP := proxyRequest("GET", profURL, nil, r.Header)
	if errP != nil {
		http.Error(w, errP.Error(), http.StatusBadGateway)
		return
	}

	// Si uno devuelve 404 → devolver 404
	if statusS == http.StatusNotFound || statusP == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		if statusS == http.StatusNotFound {
			w.Write(bodyS)
		} else {
			w.Write(bodyP)
		}
		return
	}

	// Unir JSON
	var mS map[string]interface{}
	var mP map[string]interface{}
	json.Unmarshal(bodyS, &mS)
	json.Unmarshal(bodyP, &mP)

	for k, v := range mP {
		if _, ok := mS[k]; !ok {
			mS[k] = v
		}
	}

	out, _ := json.Marshal(mS)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(out)
}

// -------------------------------------------
// PUT USER — DIVIDIR PAYLOAD + UNIFICAR RESPUESTA
// -------------------------------------------
func handleUpdateUserFull(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromEnv()
	vars := mux.Vars(r)
	id := vars["id"]

	// Leer el body
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Parsear JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Claves de cada microservicio
	securityKeys := map[string]bool{"email": true, "username": true, "password": true}
	profileKeys := map[string]bool{"firstName": true, "lastName": true, "bio": true, "avatar": true, "address": true, "phone": true}

	secPart := make(map[string]interface{})
	profPart := make(map[string]interface{})

	for k, v := range payload {
		switch {
		case securityKeys[k]:
			secPart[k] = v
		case profileKeys[k]:
			profPart[k] = v
		default:
			// Si la clave no está categorizada, enviarla a ambos
			secPart[k] = v
			profPart[k] = v
		}
	}

	// ---- PUT a SECURITY ----
	secURL := cfg.SecurityURL + "/users/" + id
	secBody, _ := json.Marshal(secPart)
	statusS, bodyS, _, errS := proxyRequest("PUT", secURL, bytes.NewReader(secBody), r.Header)
	if errS != nil {
		http.Error(w, errS.Error(), http.StatusBadGateway)
		return
	}

	// ---- PUT a PROFILE ----
	profURL := cfg.ProfileURL + "/profiles/" + id
	profBody, _ := json.Marshal(profPart)
	statusP, bodyP, _, errP := proxyRequest("PUT", profURL, bytes.NewReader(profBody), r.Header)
	if errP != nil {
		http.Error(w, errP.Error(), http.StatusBadGateway)
		return
	}

	// Si ambos ok → unir respuesta
	if statusS < 300 && statusP < 300 {
		var mS map[string]interface{}
		var mP map[string]interface{}
		json.Unmarshal(bodyS, &mS)
		json.Unmarshal(bodyP, &mP)

		for k, v := range mP {
			if _, ok := mS[k]; !ok {
				mS[k] = v
			}
		}

		out, _ := json.Marshal(mS)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(out)
		return
	}

	// Si uno falla, devolver el error del primero
	if statusS >= 400 {
		w.WriteHeader(statusS)
		w.Write(bodyS)
		return
	}

	w.WriteHeader(statusP)
	w.Write(bodyP)
}

// -------------------------------------------
// UTILIDAD PARA COPIAR HEADERS
// -------------------------------------------
func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
