package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
<<<<<<< HEAD
	"time"
=======
>>>>>>> main

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/fallback"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/go-chi/chi/v5"
	slogchi "github.com/samber/slog-chi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	r := chi.NewRouter()
	r.Use(slogchi.New(logger))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Message string `json:"message"`
		}
		// Create a Timeout for 1 second
		timeout := newTimeout(logger)

		// Use the Timeout with a failsafe RoundTripper
		roundTripper := failsafehttp.NewRoundTripper(nil, timeout)
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

func newTimeout(logger *slog.Logger) fallback.Fallback[*http.Response] {
	return timeout.Builder[*http.Response](1 * time.Second).
		OnTimeoutExceeded(func(e failsafe.ExecutionDoneEvent[*http.Response]) {
			logger.Info("Connection timed out")
		}).Build()
}
