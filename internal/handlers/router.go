package handlers

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
)

// New creates a router and registers all the routes for the
// service and returns it.
func GetRouter() http.Handler {
	// Create Router
	router := httprouter.New()

	setPingRoutes(router)

	return router
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("HTTP request sent to %s from %s", r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// 1. Ip address vice rateLimmiting
// 2. after certain time we clear our map with inactive ips
func RateLimiterWithIP(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// after 1 min, time we clear our map with inactive ips, with 1 Minute
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 2*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Println("rateLimitExceed")
			return
		} else {
			fmt.Println(" Request allowed ")
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter:  rate.NewLimiter(2, 4),
				lastSeen: time.Now(),
			}
		}

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			fmt.Println("rateLimitExceed")
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

/*
func Middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Printf("HTTP request sent to %s from %s", r.URL.Path, r.RemoteAddr)
		// call registered handler
		n(w, r, ps)
	}
}

func RateLimiter(next http.Handler) http.Handler {
	limit := rate.NewLimiter(2, 4) //2 req-rate/sec , maxRequest=4
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limit.Allow() {
			fmt.Println(" rateLimitExceed ")
		} else {
			fmt.Println(" Request allowed ")
		}
		next.ServeHTTP(w, r)
	})
}
*/
