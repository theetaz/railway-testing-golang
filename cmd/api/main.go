package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
)

type ServerSpecs struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	NumCPU       int    `json:"num_cpu"`
	GoVersion    string `json:"go_version"`
}

func main() {
	http.HandleFunc("/", getServerSpecs)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getServerSpecs(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	specs := ServerSpecs{
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GoVersion:    runtime.Version(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(specs)
}