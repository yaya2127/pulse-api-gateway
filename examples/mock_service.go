package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	port := flag.Int("port", 9001, "Mock service port")
	name := flag.String("name", "UserService", "Service name")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"service":   *name,
			"path":      r.URL.Path,
			"method":    r.Method,
			"headers":   r.Header,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		}
		json.NewEncoder(w).Encode(resp)
	})

	log.Printf("Starting %s on :%d", *name, *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)
}
