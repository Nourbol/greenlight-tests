package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("mocked error"))
	})

	handlerToTest := app.recoverPanic(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	expectedConnectionHeaderValue := "close"
	actualConnectionHeaderValue := rr.Header().Get("Connection")

	if actualConnectionHeaderValue != expectedConnectionHeaderValue {
		t.Errorf(
			"expected 'Connection' header with %s value, but got %s",
			expectedConnectionHeaderValue, actualConnectionHeaderValue)
	}

	expectedStatusCode := http.StatusInternalServerError

	if rr.Code != expectedStatusCode {
		t.Errorf("expected %d status code but got %d", expectedStatusCode, rr.Code)
	}
}

func TestRateLimit(t *testing.T) {
	app := newTestApplication(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handlerToTest := app.rateLimit(nextHandler)

	tests := []struct {
		name         string
		enabled      bool
		rps          float64
		burst        int
		expectedCode int
		handlerCalls int
		hostname     string
	}{
		{
			name:         "Exceeding burst of a client",
			enabled:      true,
			handlerCalls: 2,
			rps:          4,
			burst:        1,
			expectedCode: http.StatusTooManyRequests,
		},
		{
			name:         "Sending enough requests",
			enabled:      true,
			handlerCalls: 2,
			rps:          2,
			burst:        4,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid hostname",
			enabled:      true,
			handlerCalls: 1,
			rps:          1,
			burst:        1,
			expectedCode: http.StatusInternalServerError,
			hostname:     "invalid_hostname-",
		},
		{
			name:         "Disabled rate limiter",
			enabled:      false,
			handlerCalls: 2,
			rps:          1,
			burst:        1,
			expectedCode: http.StatusOK,
		},
	}

	for i, e := range tests {

		req := httptest.NewRequest("GET", "http://testing", nil)

		if e.hostname == "" {
			e.hostname = ":2000"
		}
		req.RemoteAddr = fmt.Sprintf("%d%s", i, e.hostname)

		rr := httptest.NewRecorder()
		app.config.limiter = struct {
			rps     float64
			burst   int
			enabled bool
		}{
			rps:     e.rps,
			burst:   e.burst,
			enabled: e.enabled,
		}

		var lastStatusCode int

		for i := 0; i < e.handlerCalls; i++ {
			handlerToTest.ServeHTTP(rr, req)
			lastStatusCode = rr.Code
		}

		if e.expectedCode != lastStatusCode {
			t.Errorf("%s: there is no response with %d code", e.name, e.expectedCode)
		}
	}
}
