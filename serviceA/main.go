package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/bulkhead"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		bulkhead := bulkhead.Builder[*http.Response](1).
			WithMaxWaitTime(1 * time.Second).
			OnFull(func(e failsafe.ExecutionEvent[*http.Response]) {
				fmt.Println("Bulkhead full")
				fmt.Println(e)
				panic("Bulkhead full")
			}).
			Build()

		roundTripper := failsafehttp.NewRoundTripper(nil, bulkhead)
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
		type response struct {
			Message string `json:"message"`
		}
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
