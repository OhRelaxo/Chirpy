package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/OhRelaxo/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	devMode        bool
}

func main() {
	const port = "8080"
	const rootPath = "."

	godotenv.Load()

	devMode := false
	if dev := os.Getenv("PLATFORM"); dev == "Dev" {
		devMode = true
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, devMode: devMode}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(rootPath)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

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
		log.Printf("error in <handlerReadiness> at w.Write: %v", err)
	}
}
