// Command demoserver supplies invented API responses for an account-free VHS demo.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := "127.0.0.1:8645"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           http.HandlerFunc(serveDemo),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.Println("invented demo API listening")
	log.Fatal(server.ListenAndServe())
}

func serveDemo(w http.ResponseWriter, r *http.Request) {
	var result any
	switch r.Method + " " + r.URL.Path {
	case "GET /conversations.list":
		result = map[string]any{"ok": true, "channels": []any{
			map[string]any{"id": "CDEMO001", "name": "demo-builds", "is_private": false, "is_archived": false, "num_members": 12},
			map[string]any{"id": "CDEMO002", "name": "demo-help", "is_private": false, "is_archived": false, "num_members": 8},
		}, "response_metadata": map[string]any{"next_cursor": ""}}
	case "GET /conversations.history":
		result = map[string]any{"ok": true, "has_more": false, "messages": []any{
			map[string]any{"type": "message", "user": "UDEMO001", "text": "The demo build passed all checks.", "ts": "1790762400.000100"},
			map[string]any{"type": "message", "user": "UDEMO002", "text": "Ready to ship the invented release.", "ts": "1790762300.000200"},
		}, "response_metadata": map[string]any{"next_cursor": ""}}
	case "POST /chat.postMessage":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		result = map[string]any{"ok": true, "channel": r.Form.Get("channel"), "ts": "1790762500.000300", "message": map[string]any{"type": "message", "user": "UDEMOBOT", "text": r.Form.Get("text"), "ts": "1790762500.000300"}}
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Print(err)
	}
}
