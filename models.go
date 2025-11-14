package main

// ...existing code...
// Tipos auxiliares mínimos
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
