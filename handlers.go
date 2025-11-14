package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

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

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromEnv()
	vars := mux.Vars(r)
	id := vars["id"]
	target := cfg.SecurityURL + "/users/" + id

	// Reenviar la eliminación al servicio de seguridad
	status, body, headers, err := proxyRequest("DELETE", target, nil, r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Si la eliminación fue exitosa (2xx), publicar evento user.deleted
	if status >= 200 && status < 300 {
		event := map[string]interface{}{
			"type": "user.deleted",
			"payload": map[string]string{
				"userId": id,
			},
		}
		if err := postEvent(cfg, event); err != nil {
			// No fallar la operación principal; sólo loguear
			log.Printf("falló publicar evento de eliminación: %v", err)
		}
	}

	copyHeaders(w.Header(), headers)
	w.WriteHeader(status)
	w.Write(body)
}

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

	// Si uno de los dos devolvió no encontrado, propagar
	if statusS == http.StatusNotFound || statusP == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		// intentar devolver el body del servicio que indicó 404
		if statusS == http.StatusNotFound {
			w.Write(bodyS)
		} else {
			w.Write(bodyP)
		}
		return
	}

	// Unir ambos JSON en uno solo; si hay conflictos, prevalece security
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

func handleUpdateUserFull(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromEnv()
	vars := mux.Vars(r)
	id := vars["id"]

	// Leer body completo
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Dividir el JSON en dos partes basadas en claves conocidas
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	securityKeys := map[string]bool{"email": true, "username": true, "password": true}
	profileKeys := map[string]bool{"firstName": true, "lastName": true, "bio": true, "avatar": true, "address": true, "phone": true}

	secPart := make(map[string]interface{})
	profPart := make(map[string]interface{})

	for k, v := range payload {
		if securityKeys[k] {
			secPart[k] = v
		} else if profileKeys[k] {
			profPart[k] = v
		} else {
			// Si la clave no está categorizada, enviarla a ambos por seguridad
			secPart[k] = v
			profPart[k] = v
		}
	}

	// Enviar a security
	secURL := cfg.SecurityURL + "/users/" + id
	secBody, _ := json.Marshal(secPart)
	statusS, bodyS, _, errS := proxyRequest("PUT", secURL, bytes.NewReader(secBody), r.Header)
	if errS != nil {
		http.Error(w, errS.Error(), http.StatusBadGateway)
		return
	}

	// Enviar a profile
	profURL := cfg.ProfileURL + "/profiles/" + id
	profBody, _ := json.Marshal(profPart)
	statusP, bodyP, _, errP := proxyRequest("PUT", profURL, bytes.NewReader(profBody), r.Header)
	if errP != nil {
		http.Error(w, errP.Error(), http.StatusBadGateway)
		return
	}

	// Unir respuestas (si ambos OK, 200; si alguno falla, propagar)
	if statusS >= 200 && statusS < 300 && statusP >= 200 && statusP < 300 {
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

	// Si hubo error en alguno, devolver el primero con error >=400
	if statusS >= 400 {
		w.WriteHeader(statusS)
		w.Write(bodyS)
		return
	}
	w.WriteHeader(statusP)
	w.Write(bodyP)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
