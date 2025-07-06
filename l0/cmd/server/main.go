package main

import (
	"context"
	"log"
	"l0/internal/api"
	"l0/internal/cache"
	"l0/internal/kafka"
	"l0/internal/storage"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация БД
	if err := storage.InitDB(ctx, "postgres://orders_user:orders_password@localhost:5432/orders_db"); err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	// Инициализация кэша
	c := cache.NewCache()

	// Восстановление данных из БД
	orders, err := storage.LoadOrders(ctx)
	if err != nil {
		log.Fatal("Failed to load orders:", err)
	}
	for _, order := range orders {
		c.Set(order)
	}
	log.Printf("Restored %d orders from DB", len(orders))

	// Запуск Kafka consumer
	go kafka.StartConsumer(c) // Исправленный вызов

	// Запуск HTTP сервера
	srv := api.NewServer(c)
	go func() {
		if err := srv.Start(":8080"); err != nil {
			log.Fatal("HTTP server error:", err)
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")
	cancel()
}