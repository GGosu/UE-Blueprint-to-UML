package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"UE_UML/internal/blueprint"
	"UE_UML/templates"
)

//go:embed static
var staticFS embed.FS

func main() {
	cfg, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("cannot load config.yml: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", makeIndexHandler(cfg))
	mux.HandleFunc("POST /convert", makeConvertHandler(cfg))
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("GET /static/", http.FileServerFS(staticFS))

	log.Printf("Listening on http://localhost:%d", cfg.Port)
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		log.Fatal(err)
	}
}

func makeIndexHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Index(cfg.Version).Render(r.Context(), w)
	}
}

func makeConvertHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyKB*1024)
		convertHandler(w, r)
	}
}

type convertResponse struct {
	Mermaid string `json:"mermaid,omitempty"`
	Error   string `json:"error,omitempty"`
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	writeErr := func(msg string) {
		json.NewEncoder(w).Encode(convertResponse{Error: msg})
	}

	bp := strings.TrimSpace(r.FormValue("blueprint"))
	if bp == "" {
		writeErr("No Blueprint text received.")
		return
	}

	g, err := blueprint.ParseBlueprint(bp)
	if err != nil {
		writeErr(fmt.Sprintf("Parse error: %s", err.Error()))
		return
	}
	if len(g.Nodes) == 0 {
		writeErr("No Blueprint nodes found. Copy nodes in UE Editor with Ctrl+C first.")
		return
	}

	json.NewEncoder(w).Encode(convertResponse{Mermaid: blueprint.GenerateMermaid(g)})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
