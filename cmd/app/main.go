package main

import (
 "context"
 "l0/internal/api"
 "l0/internal/cache"
 "l0/internal/kafka"
 "l0/internal/storage"
 "log"
)

func main() {
 ctx := context.Background()

 // подключаем БД
 if err := storage.InitDB(ctx, "postgres://orders_user:orders_password@localhost:5432/orders_db"); err != nil {
  log.Fatal(err)
 }

 // создаём кэш
 c := cache.NewCache()

 // восстанавливаем данные из БД в кэш при старте
 // восстановление кэша из базы
 orders, err := storage.LoadOrders(ctx)
 if err != nil {
	log.Fatal("failed to load orders from db:", err)
 }
 for _, order := range orders {
	c.Set(order)
 } 
log.Println("Cache initialized")


 // запускаем консьюмера
 go kafka.StartConsumer(c)

 // запускаем API
 srv := api.Server{Cache: c}
 srv.Start()
}