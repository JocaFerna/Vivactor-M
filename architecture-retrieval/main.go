package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"architecture-retrieval/routes"
	"github.com/rs/cors"
	"architecture-retrieval/last_will"
)

func main() {
	last_will.SetupCleanupHandler()
	port := env("PORT", "8000")
	handler(fmt.Sprintf(":%s", port))
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func handler(address string) error {
	routes.Register()
	log.Printf("Listening on %s", address)

	// 2. Criamos a configuração de CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, 
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	mainHandler := c.Handler(http.DefaultServeMux)

	err := http.ListenAndServe(address, mainHandler)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
