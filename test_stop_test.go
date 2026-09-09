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
		`>1m</time>`,
		`>5m</time>`,
		`>17m</time>`,
		`>32m</time>`,
	} {
		if !strings.Contains(w.Body.String(), want) {
			t.Errorf("response is missing %q", want)
		}
	}
}

func TestFormatCountdown(t *testing.T) {
	for offset, want := range map[time.Duration]string{
		90 * time.Second:          "1m",
		60500 * time.Millisecond:  "1m",
		60 * time.Second:          "60s",
		59900 * time.Millisecond:  "59s",
		0:                         "0s",
		-500 * time.Millisecond:   "0s",
		-60 * time.Second:         "-60s",
		-60500 * time.Millisecond: "-1m",
		-90 * time.Second:         "-1m",
	} {
		t.Run(offset.String(), func(t *testing.T) {
			arrival := testNow.Add(offset).Format(time.RFC3339Nano)
			if got := formatCountdown(arrival, testNow); got != want {
				t.Errorf("countdown = %q, want %q", got, want)
			}
		})
	}
	if got := formatCountdown("unknown", testNow); got != "unknown" {
		t.Errorf("invalid arrival = %q, want unknown", got)
	}
}
