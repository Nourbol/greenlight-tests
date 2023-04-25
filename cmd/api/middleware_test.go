package main

import (
	"errors"
	"fmt"
	"greenlight.bcc/internal/data"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestAuthenticate(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name                     string
		authorizationHeaderValue string
		expectedCode             int
		mustBeAnonymous          bool
	}{
		{
			name:            "Without authentication",
			expectedCode:    http.StatusOK,
			mustBeAnonymous: true,
		},
		{
			name:                     "With valid token",
			authorizationHeaderValue: "Bearer eyJhbGciO.eyJzNTY3ODI.SflK",
			expectedCode:             http.StatusOK,
		},
		{
			name:                     "Without Bearer prefix",
			authorizationHeaderValue: "eyJhbGciO.eyJzNTY3ODI.SflK",
			expectedCode:             http.StatusUnauthorized,
		},
		{
			name:                     "Token is not 26 bytes long",
			authorizationHeaderValue: "Bearer eyJhbGciO.eyJzNTY3ODI.SflKsG",
			expectedCode:             http.StatusUnauthorized,
		},
		{
			name:                     "Without Bearer prefix",
			authorizationHeaderValue: "eyJhbGciO.eyJzNTY3ODI.SflK",
			expectedCode:             http.StatusUnauthorized,
		},
		{
			name:                     "Non-existent token",
			authorizationHeaderValue: "Bearer non_ex_token.DIDgtWaS.SflK",
			expectedCode:             http.StatusUnauthorized,
		},
		{
			name:                     "Non-existent token",
			authorizationHeaderValue: "Bearer non_ex_token.unexpec_error",
			expectedCode:             http.StatusInternalServerError,
		},
	}

	for _, e := range tests {

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(userContextKey).(*data.User)
			if !ok {
				t.Fatalf("%s: no user in context", e.name)
			}
			if e.mustBeAnonymous != user.IsAnonymous() {
				t.Errorf("%s: expected user to be anonymous", e.name)
			}
		})

		handlerToTest := app.authenticate(nextHandler)

		req := httptest.NewRequest("GET", "http://testing", nil)
		req.Header.Add("Authorization", e.authorizationHeaderValue)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != e.expectedCode {
			t.Errorf("%s: expected %d but got %d", e.name, e.expectedCode, rr.Code)
		}
	}
}

func TestRequiredAuthenticatedUser(t *testing.T) {
	app := newTestApplication(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handlerToTest := app.requireAuthenticatedUser(nextHandler)

	tests := []struct {
		name         string
		user         *data.User
		expectedCode int
	}{
		{
			name:         "With the authenticated user",
			expectedCode: http.StatusOK,
			user: &data.User{
				ID:        1,
				CreatedAt: time.Now(),
				Name:      "John Doe",
				Email:     "johndoe@greenlight.test",
				Activated: true,
				Version:   0,
			},
		},
		{
			name:         "With an anonymous user",
			expectedCode: http.StatusUnauthorized,
			user:         data.AnonymousUser,
		},
	}

	for _, e := range tests {

		req := httptest.NewRequest("GET", "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, app.contextSetUser(req, e.user))

		if rr.Code != e.expectedCode {
			t.Errorf("%s: expected %d but got %d", e.name, e.expectedCode, rr.Code)
		}
	}
}

func TestRequireActivatedUser(t *testing.T) {
	app := newTestApplication(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handlerToTest := app.requireActivatedUser(nextHandler)

	tests := []struct {
		name         string
		user         *data.User
		expectedCode int
	}{
		{
			name:         "With an unactivated user",
			expectedCode: http.StatusOK,
			user: &data.User{
				Activated: true,
			},
		},
		{
			name:         "With an activated user",
			expectedCode: http.StatusForbidden,
			user: &data.User{
				Activated: false,
			},
		},
	}

	for _, e := range tests {

		req := httptest.NewRequest("GET", "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, app.contextSetUser(req, e.user))

		if rr.Code != e.expectedCode {
			t.Errorf("%s: expected %d but got %d", e.name, e.expectedCode, rr.Code)
		}
	}
}

func TestRequirePermission(t *testing.T) {
	app := newTestApplication(t)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handlerToTest := app.requirePermission("mocked:code", nextHandler)

	tests := []struct {
		name         string
		user         *data.User
		expectedCode int
	}{
		{
			name:         "User with permission",
			expectedCode: http.StatusOK,
			user: &data.User{
				ID:        1,
				Activated: true,
			},
		},
		{
			name:         "User with wrong permission",
			expectedCode: http.StatusForbidden,
			user: &data.User{
				ID:        3,
				Activated: true,
			},
		},
		{
			name:         "Unexpected error from Model",
			expectedCode: http.StatusInternalServerError,
			user: &data.User{
				ID:        2,
				Activated: true,
			},
		},
	}

	for _, e := range tests {

		req := httptest.NewRequest("GET", "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, app.contextSetUser(req, e.user))

		if rr.Code != e.expectedCode {
			t.Errorf("%s: expected %d but got %d", e.name, e.expectedCode, rr.Code)
		}
	}
}
