package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		retryAfterDelay := 1 * time.Second
		if fail() {
			w.Header().Add("Retry-After", strconv.Itoa(int(retryAfterDelay.Seconds())))
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello from service B"}`))
	})
	http.ListenAndServe(":3001", r)
}

func fail() bool {
	if flipint := rand.Intn(2); flipint == 0 {
		return true
	}
	return false
}
