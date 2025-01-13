package main

import (
	"net/http"
	"time"

	"metric"
)

func main() {

	mtr, err := metric.NewMonitor(
		metric.WithRequestTotal(),
		metric.WithSlowRequest(time.Second),
	)
	if err != nil {
		panic(err)
	}

	var handler http.Handler

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, mw := range mtr.Middlewares() {
		handler = mw(handler)
	}

	http.Handle("/metrics", mtr.Expose())
	http.Handle("/", handler)

	http.ListenAndServe(":8080", nil)
}
