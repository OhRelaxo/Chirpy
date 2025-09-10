package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const port = "8080"
	const rootPath = "."

	apiCfg := apiConfig{fileserverHits: atomic.Int32{}}

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(rootPath)))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", rootPath, port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, _ *http.Request) {
	log.Println("the server is Ready")
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("log in <handlerReadiness> at w.Write:\n%v", err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("adding hits")
		cfg.fileserverHits.Add(int32(1))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, _ *http.Request) {
	log.Println("Outputting Hits")
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load())))
	if err != nil {
		log.Printf("log in <handlerMetrics at w.Write:\n%v", err)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, _ *http.Request) {
	log.Println("Resetting Hits")
	cfg.fileserverHits.Store(int32(0))
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Printf("log in <middlewareMetricsReset> at w.Write:\n%v", err)
	}
}
