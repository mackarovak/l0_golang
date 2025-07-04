package api

import (
	"encoding/json"
	"log"
	"net/http"

	"l0/internal/cache"
)

type Server struct {
	Cache *cache.Cache
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/order/", s.getOrderHandler)

	addr := ":8081"
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func (s *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Path[len("/order/"):]
	if orderID == "" {
		http.Error(w, "order_id is required", http.StatusBadRequest)
		return
	}

	order, ok := s.Cache.Get(orderID)
	if !ok {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "failed to encode json", http.StatusInternalServerError)
	}
}
