package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"level_zero/config"
	"level_zero/internal/cache"
	"level_zero/internal/db"
	"level_zero/internal/handler"

	"github.com/nats-io/stan.go"

	"github.com/gorilla/mux"
)

func main() {
	configPath := "./config/config.json"
	schemaPath := "./internal/db/scripts/create_schema.sql"
	config := config.InitialzeConfig(configPath)

	conn, err := db.InitialzeDatabase(config, schemaPath)
	if err != nil {
		log.Fatal("Database initalization error: ", err)
	}
	defer conn.Close()
	log.Print("Successfuly initialazed database")

	cache := cache.CreateCache()
	err = cache.InitialzeCache(conn)
	if err != nil {
		log.Print("Cache initialization error: ", err)
	}
	log.Print("Cache initialazed with ", len(cache.Data), " orders")

	natsUrl := fmt.Sprintf("nats://%s", config.NatsUrl)
	nc, err := stan.Connect(config.NatsCluster, "subscriber", stan.NatsURL(natsUrl))
	if err != nil {
		log.Fatal("Nats connection error: ", err)
	}
	defer nc.Close()

	nc.Subscribe(config.NatsSubject, handler.MessageHandler(*cache, conn))

	router := mux.NewRouter()
	router.HandleFunc("/", handler.IndexHandler)
    router.HandleFunc("/order/{uid:[a-z0-9]+}", handler.OrderHandler(*cache))
    http.Handle("/", router)
	log.Printf("Server is running on http://localhost:%s", config.ServerPort)
	port := fmt.Sprintf(":%s", config.ServerPort)
    err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Server listengng and serving error: ", err)
	}
}