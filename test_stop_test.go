package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 8, 24, 21, 0, 0, 0, time.FixedZone("SGT", 8*60*60))

func TestTestArrivals(t *testing.T) {
	arrivals := testArrivals(testNow)

	if arrivals.BusStopCode != testStopCode {
		t.Errorf("BusStopCode = %q, want %q", arrivals.BusStopCode, testStopCode)
	}
	if len(arrivals.Services) != 2 {
		t.Fatalf("got %d services, want 2", len(arrivals.Services))
	}
	if got := arrivals.Services[0].ServiceNo; got != "42" {
		t.Errorf("first service = %q, want 42", got)
	}
	if got, want := arrivals.Services[0].NextBus.EstimatedArrival, "2026-08-24T21:01:30+08:00"; got != want {
		t.Errorf("first arrival = %q, want %q", got, want)
	}
}

func TestTestStop(t *testing.T) {
	server, err := NewServer("static/all.json")
	if err != nil {
		t.Fatal(err)
	}
	server.now = func() time.Time { return testNow }

	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?id="+testStopCode, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	for _, want := range []string{
		testStopName,
		`name="robots" content="noindex"`,
		`class="load-seats"`,
		`class="load-standing"`,
		`class="load-full"`,
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Errorf("response is missing %q", want)
		}
	}
}
