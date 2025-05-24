package main

import (
	"net/http"
)

func main() {
	dir := http.Dir(".")
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(dir))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidation)

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
