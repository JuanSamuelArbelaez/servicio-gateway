package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// HttpClient exportado para que main pueda configurarlo
var HttpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// ProxyRequest envía la petición al servicio objetivo y devuelve status, body, headers
func ProxyRequest(method, url string, body io.Reader, headers http.Header) (int, []byte, http.Header, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, nil, err
	}

	// Copiar headers (saltear Host)
	for k, vv := range headers {
		if k == "Host" {
			continue
		}
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Printf("[client] error calling %s: %v\n", url, err)
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}

	return resp.StatusCode, respBody, resp.Header, nil
}

// PostEvent publica un evento en el EventBus usando directamente la URL.
// eventBusURL: la URL base del event-bus, por ejemplo "http://notification-orchestrator:8085" o "http://notification-orchestrator:8085/api"
// event: cualquier estructura serializable a JSON
func PostEvent(eventBusURL string, event interface{}) error {
	if eventBusURL == "" {
		log.Printf("[client] EventBus not configured, skipping event: %+v\n", event)
		return nil
	}

	// Normalizar URL y construir endpoint /events
	target := eventBusURL
	// remove trailing slash
	if len(target) > 0 && target[len(target)-1] == '/' {
		target = target[:len(target)-1]
	}
	target = target + "/events"

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Printf("[client] error posting event to %s: %v\n", target, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[client] event-bus returned status %d: %s\n", resp.StatusCode, string(b))
	}

	return nil
}
