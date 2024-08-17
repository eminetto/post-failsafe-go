package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/fallback"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     map[string][]string{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewBufferString(`{"message": "error accessing service B"}`)),
		}
		fallback := fallback.BuilderWithResult[*http.Response](resp).
			HandleIf(func(response *http.Response, err error) bool {
				return response != nil && response.StatusCode == http.StatusServiceUnavailable
			}).
			OnFallbackExecuted(func(e failsafe.ExecutionDoneEvent[*http.Response]) {
				fmt.Println("Fallback executed result")
			}).
			Build()

		roundTripper := failsafehttp.NewRoundTripper(nil, fallback)
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
