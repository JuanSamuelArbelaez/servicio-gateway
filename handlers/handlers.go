package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"servicio-gateway/client"
	"servicio-gateway/config"
)

// MAKE PROXY FOR SECURITY SERVICE
func MakeProxyToSecurity(method, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cfg := config.LoadConfigFromEnv()

		// Build dynamic URL
		target := strings.TrimRight(cfg.SecurityURL, "/") + path

		// Replace path vars: {id}
		for k, v := range mux.Vars(r) {
			target = strings.Replace(target, "{"+k+"}", v, 1)
		}

		status, body, headers, err := client.ProxyRequest(method, target, r.Body, r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		CopyHeaders(w.Header(), headers)
		w.WriteHeader(status)
		w.Write(body)
	}
}

// DELETE USER → SEND EVENT user.deleted
func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfigFromEnv()
	id := mux.Vars(r)["id"]

	target := strings.TrimRight(cfg.SecurityURL, "/") + "/api/v1/users/" + id

	status, body, headers, err := client.ProxyRequest("DELETE", target, nil, r.Header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// If deleted successfully → publish event (pass URL)
	if status >= 200 && status < 300 {
		event := map[string]interface{}{
			"type": "user.deleted",
			"payload": map[string]interface{}{
				"userId": id,
			},
		}
		_ = client.PostEvent(cfg.EventBusURL, event)
	}

	CopyHeaders(w.Header(), headers)
	w.WriteHeader(status)
	w.Write(body)
}

// GET USER FULL → MERGE SECURITY + PROFILE
func HandleGetUserFull(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfigFromEnv()
	id := mux.Vars(r)["id"]

	// SECURITY USER
	secURL := strings.TrimRight(cfg.SecurityURL, "/") + "/api/v1/users/" + id
	statusS, bodyS, _, errS := client.ProxyRequest("GET", secURL, nil, r.Header)
	if errS != nil {
		http.Error(w, errS.Error(), http.StatusBadGateway)
		return
	}

	// PROFILE USER
	profURL := strings.TrimRight(cfg.ProfileURL, "/") + "/api/v1/profiles/" + id
	statusP, bodyP, _, errP := client.ProxyRequest("GET", profURL, nil, r.Header)
	if errP != nil {
		http.Error(w, errP.Error(), http.StatusBadGateway)
		return
	}

	if statusS == http.StatusNotFound || statusP == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		if statusS == http.StatusNotFound {
			w.Write(bodyS)
		} else {
			w.Write(bodyP)
		}
		return
	}

	var mS, mP map[string]interface{}
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

// UPDATE USER FULL → SPLIT DATA INTO SECURITY + PROFILE
func HandleUpdateUserFull(w http.ResponseWriter, r *http.Request) {
	cfg := config.LoadConfigFromEnv()
	id := mux.Vars(r)["id"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	securityKeys := map[string]bool{
		"email": true, "username": true, "password": true,
	}
	profileKeys := map[string]bool{
		"firstName": true, "lastName": true, "bio": true,
		"avatar": true, "address": true, "phone": true,
	}

	secPart := map[string]interface{}{}
	profPart := map[string]interface{}{}

	for k, v := range payload {
		if securityKeys[k] {
			secPart[k] = v
		} else if profileKeys[k] {
			profPart[k] = v
		} else {
			secPart[k] = v
			profPart[k] = v
		}
	}

	// SECURITY UPDATE
	secURL := strings.TrimRight(cfg.SecurityURL, "/") + "/api/v1/users/" + id
	secBody := jsonMarshal(secPart)
	statusS, bodyS, _, errS := client.ProxyRequest("PUT", secURL, bytes.NewReader(secBody), r.Header)
	if errS != nil {
		http.Error(w, errS.Error(), http.StatusBadGateway)
		return
	}

	// PROFILE UPDATE
	profURL := strings.TrimRight(cfg.ProfileURL, "/") + "/api/v1/profiles/" + id
	profBody := jsonMarshal(profPart)
	statusP, bodyP, _, errP := client.ProxyRequest("PUT", profURL, bytes.NewReader(profBody), r.Header)
	if errP != nil {
		http.Error(w, errP.Error(), http.StatusBadGateway)
		return
	}

	if statusS >= 200 && statusS < 300 && statusP >= 200 && statusP < 300 {
		var mS, mP map[string]interface{}
		json.Unmarshal(bodyS, &mS)
		json.Unmarshal(bodyP, &mP)
		for k, v := range mP {
			if _, ok := mS[k]; !ok {
				mS[k] = v
			}
		}
		out := jsonMarshal(mS)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(out)
		return
	}

	if statusS >= 400 {
		w.WriteHeader(statusS)
		w.Write(bodyS)
		return
	}
	w.WriteHeader(statusP)
	w.Write(bodyP)
}

// UTILS
func CopyHeaders(dst, src http.Header) {
	for k, v := range src {
		for _, h := range v {
			dst.Add(k, h)
		}
	}
}

func jsonMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
