package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	responseTimeHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "response_time_milliseconds",
		Help:    "The histogram of response time in milliseconds",
		Buckets: []float64{.5, 1, 2.5, 5, 10, 25, 50, 100, 150, 200},
	})
	processedRequestsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "processed_requests_total",
		Help: "The total count of processed requests",
	}, []string{"status_code"})
)
