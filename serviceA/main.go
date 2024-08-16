package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
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
		// Create a RetryPolicy that only handles 500 responses, with backoff delays between retries
		retryPolicy := retrypolicy.Builder[*http.Response]().
			HandleIf(func(response *http.Response, _ error) bool {
				return response != nil && response.StatusCode == 500
			}).
			WithBackoff(time.Second, 10*time.Second).
			OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[*http.Response]) {
				fmt.Println("Retry", e.Attempts(), "after delay of", e.Delay)
			}).Build()

		// Use the RetryPolicy with a failsafe RoundTripper
		roundTripper := failsafehttp.NewRoundTripper(nil, retryPolicy)
		client := &http.Client{Transport: roundTripper}

		resp, err := client.Get("http://localhost:3001")
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
