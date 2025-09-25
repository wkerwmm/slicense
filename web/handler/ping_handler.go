package handler

import (
    "encoding/json"
    "net/http"
    "time"
)

var startTime = time.Now()

type PingResponse struct {
    Message    string            `json:"message"`
    Version    string            `json:"version"`
    Uptime     string            `json:"uptime"`
    Timestamp  string            `json:"timestamp"`
    Method     string            `json:"method"`
    ClientIP   string            `json:"client_ip"`
    Headers    map[string]string `json:"headers"`
    Health     string            `json:"health"`
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.URL.Query().Get("error") == "true" {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "simulated server error",
        })
        return
    }

    uptime := time.Since(startTime).Round(time.Second).String()

    clientIP := r.RemoteAddr
    if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
        clientIP = forwarded
    }

    headers := make(map[string]string)
    for name, values := range r.Header {
        if len(values) > 0 {
            headers[name] = values[0]
        }
    }

    response := PingResponse{
        Message:   "Pong",
        Version:   "v1.2.3",
        Uptime:    uptime,
        Timestamp: time.Now().Format(time.RFC3339),
        Method:    r.Method,
        ClientIP:  clientIP,
        Headers:   headers,
        Health:    "OK",
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
