package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go/circuitbreaker"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Message string `json:"message"`
		}
		// Create a CircuitBreaker that handles 503 responses and uses a half-open delay based on the Retry-After header
		circuitBreaker := circuitbreaker.Builder[*http.Response]().
			HandleIf(func(response *http.Response, err error) bool {
				return response != nil && response.StatusCode == http.StatusServiceUnavailable
			}).
			WithDelayFunc(failsafehttp.DelayFunc).
			OnStateChanged(func(event circuitbreaker.StateChangedEvent) {
				fmt.Println("circuit breaker state changed", event)
			}).
			Build()

		// Use the RetryPolicy with a failsafe RoundTripper
		roundTripper := failsafehttp.NewRoundTripper(nil, circuitBreaker)
		client := &http.Client{Transport: roundTripper}

		sendGet := func() (*http.Response, error) {
			fmt.Println("Sending request")
			resp, err := client.Get("http://localhost:3001")
			return resp, err
		}
		maxRetries := 3
		resp, err := sendGet()
		for i := 0; i < maxRetries; i++ {
			if err == nil && resp != nil && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusTooManyRequests {
				break
			}
			time.Sleep(circuitBreaker.RemainingDelay()) // Wait for circuit breaker's delay, provided by the Retry-After header
			resp, err = sendGet()
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer resp.Body.Close()
		var data response
		err = json.Unmarshal(body, &data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"messageA": "hello from service A","messageB": "` + data.Message + `"}`))
	})
	http.ListenAndServe(":3000", r)
}
