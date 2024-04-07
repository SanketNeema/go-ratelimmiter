package main

import (
	"RateLimiterWithMiddleware/internal/handlers"
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", handlers.LogRequest(handlers.RateLimiterWithIP(handlers.GetRouter())))
	if err != nil {
		log.Fatal("error ", err.Error())
	}
}
