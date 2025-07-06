package api

import (
	"encoding/json"
	"log"
	"net/http"
	"l0/internal/cache"
)

type Server struct {
	cache *cache.Cache
}

func NewServer(c *cache.Cache) *Server {
	return &Server{cache: c}
}

func (s *Server) Start(addr string) error {
	// Создаем роутер с CORS middleware
	mux := http.NewServeMux()
	mux.Handle("/order/", enableCORS(http.HandlerFunc(s.getOrderHandler)))

	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// Middleware для CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обрабатываем OPTIONS запрос
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем запрос дальше по цепочке
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Path[len("/order/"):]
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, exists := s.cache.Get(orderID)
	if !exists {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Failed to encode order: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}