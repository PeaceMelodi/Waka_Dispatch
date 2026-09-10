package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/PeaceMelodi/waka-dispatch/internal/delivery"
	"github.com/PeaceMelodi/waka-dispatch/internal/dispatch"
	"github.com/PeaceMelodi/waka-dispatch/internal/order"
	"github.com/PeaceMelodi/waka-dispatch/internal/retry"
	"github.com/PeaceMelodi/waka-dispatch/internal/rider"
	"github.com/PeaceMelodi/waka-dispatch/internal/ws"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("connected to database successfully")

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)

	riderRepo := rider.NewRepository(db)
	riderService := rider.NewService(riderRepo, hub)
	riderHandler := rider.NewHandler(riderService)

	matcher := dispatch.NewMatcher(riderRepo)

	orderRepo := order.NewRepository(db)

	deliveryRepo := delivery.NewRepository(db)
	deliveryService := delivery.NewService(deliveryRepo, orderRepo, riderRepo, hub)
	deliveryHandler := delivery.NewHandler(deliveryService)

	orderService := order.NewService(orderRepo, matcher, riderService, deliveryRepo, hub)
	orderHandler := order.NewHandler(orderService)

	
	retryWorker := retry.NewWorker(orderRepo, riderRepo, deliveryRepo, matcher, hub)
	retryWorker.Start(context.Background())

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "index.html"))
	})

	mux.HandleFunc("POST /riders", riderHandler.Register)
	mux.HandleFunc("PATCH /riders/{id}/location", riderHandler.UpdateLocation)
	mux.HandleFunc("PATCH /riders/{id}/status", riderHandler.UpdateStatus)
	mux.HandleFunc("GET /riders", riderHandler.List)
	mux.HandleFunc("GET /riders/available", riderHandler.ListAvailable)

	mux.HandleFunc("POST /orders", orderHandler.Create)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetByID)
	mux.HandleFunc("GET /orders", orderHandler.List)
	mux.HandleFunc("PATCH /orders/{id}/cancel", orderHandler.Cancel)

	mux.HandleFunc("PATCH /deliveries/{id}/status", deliveryHandler.UpdateStatus)
	mux.HandleFunc("GET /deliveries/{id}", deliveryHandler.GetByID)
	mux.HandleFunc("GET /deliveries/active", deliveryHandler.ListActive)

	mux.HandleFunc("GET /ws/dispatch", wsHandler.HandleConnection)

	log.Printf("starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}