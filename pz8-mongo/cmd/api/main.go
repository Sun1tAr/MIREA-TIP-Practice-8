package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"example.com/pz8-mongo/internal/db"
	"example.com/pz8-mongo/internal/notes"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	uri := getenv("MONGO_URI", "mongodb://root:secret@45.150.8.109:27017/?authSource=admin")
	dbName := getenv("MONGO_DB", "pz8")
	addr := getenv("HTTP_ADDR", ":8080")

	deps, err := db.ConnectMongo(context.Background(), uri, dbName)
	if err != nil { log.Fatal("mongo connect:", err) }
	defer deps.Client.Disconnect(context.Background())

	repo, err := notes.NewRepo(deps.Database)
	if err != nil { log.Fatal("notes repo:", err) }
	h := notes.NewHandler(repo)

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Mount("/api/v1/notes", h.Routes())

	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
