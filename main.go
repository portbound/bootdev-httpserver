package main

import "net/http"

func main() {
	dir := http.Dir(".")

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(dir)))
	mux.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
