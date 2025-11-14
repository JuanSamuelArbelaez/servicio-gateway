package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var httpClient *http.Client

type Config struct {
	SecurityURL string
	ProfileURL  string
	EventBusURL string
	Port        string
}

func loadConfigFromEnv() Config {
	cfg := Config{
		SecurityURL: os.Getenv("SECURITY_URL"),
		ProfileURL:  os.Getenv("PROFILE_URL"),
		EventBusURL: os.Getenv("EVENT_BUS_URL"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg
}

func proxyRequest(method, targetURL string, body io.Reader, headers http.Header) (int, []byte, http.Header, error) {
	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return 0, nil, nil, err
	}
	// Copiar encabezados útiles
	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, resp.Header, nil
}

func postEvent(cfg Config, event interface{}) error {
	if cfg.EventBusURL == "" {
		// Si no hay bus configurado, sólo loguear
		log.Printf("EVENT_BUS_URL no configurado, evento omitido: %+v", event)
		return nil
	}
	url := cfg.EventBusURL + "/events"
	j, _ := json.Marshal(event)
	req, err := http.NewRequest("POST", url, bytes.NewReader(j))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("postEvent fallo: status=%d body=%s", resp.StatusCode, string(b))
	}
	return nil
}
