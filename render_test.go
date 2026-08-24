package main

import (
	"flag"
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update the .golden render files")

// newTestServer serves the arrivals in testdata/<fixture>.json, or an error if
// fixture is empty, so the render never depends on datamall being up.
func newTestServer(t *testing.T, fixture string) *Server {
	t.Helper()
	s, err := NewServer("static/all.json")
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s.arrivals = func(stopID string) (SGBusArrivals, error) {
		if fixture == "" {
			return SGBusArrivals{}, os.ErrDeadlineExceeded
		}
		return fixtureArrivals(filepath.Join("testdata", fixture+".json"))(stopID)
	}
	return s
}

// TestRender renders the whole page against fixtures and diffs it against a
// golden file. Run `go test -update` after intentional HTML changes and read
// the git diff to review what moved.
func TestRender(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		fixture string
		status  int
	}{
		{name: "index", url: "/", status: http.StatusOK},
		{name: "buses", url: "/?id=01019", fixture: "buses", status: http.StatusOK},
		{name: "datamalldown", url: "/?id=01019", status: http.StatusFailedDependency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			newTestServer(t, tt.fixture).mux.ServeHTTP(w, httptest.NewRequest("GET", tt.url, nil))

			if w.Code != tt.status {
				t.Errorf("status = %d, want %d", w.Code, tt.status)
			}

			golden := filepath.Join("testdata", tt.name+".golden")
			if *update {
				if err := os.WriteFile(golden, w.Body.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run: go test -update)", err)
			}
			if got := w.Body.String(); got != string(want) {
				t.Errorf("%s render differs from %s, run: go test -update", tt.url, golden)
			}
		})
	}
}

// TestRenderBuses spells out what the buses render must keep doing, so a
// regression says what broke rather than just "the golden moved".
func TestRenderBuses(t *testing.T) {
	w := httptest.NewRecorder()
	newTestServer(t, "buses").mux.ServeHTTP(w, httptest.NewRequest("GET", "/?id=01019", nil))
	body := w.Body.String()

	for _, want := range []string{
		"<title>Singapore bus arrival times for stop 01019</title>",
		"Bras Basah Cplx",                          // nameBusStop func is wired up
		`href='/icon?stop=01019'`,                  // favicon per stop
		`content="20"`,                             // page refreshes itself
		`class="load-seats"`,                       // loadClass SEA
		`class="load-standing"`,                    // loadClass SDA
		`class="load-full"`,                        // loadClass LSD
		`dateTime="2026-08-24T21:03:12&#43;08:00"`, // arrival times, + escaped by html/template
	} {
		if !strings.Contains(body, want) {
			t.Errorf("render is missing %s", want)
		}
	}

	// soonest bus first
	if i, j := strings.Index(body, ">56<"), strings.Index(body, ">857<"); i == -1 || i > j {
		t.Errorf("services not sorted by arrival: 56 at %d, 857 at %d", i, j)
	}
}

// timeRe matches how a browser sees a <time>: attribute names are
// case-insensitive, so the template's dateTime is app.js's datetime.
var timeRe = regexp.MustCompile(`(?is)<time[^>]*\sdatetime="([^"]*)"[^>]*>(.*?)</time>`)

// TestCountdown pins the contract static/app.js depends on to turn arrival
// times into a live countdown:
//
//	new Date(timings[i].getAttribute("datetime"))
//
// Rendering the right times is not enough - if they stop being <time>
// elements, or the timestamp stops being parseable, the page still looks fine
// and the countdown silently dies. That is the regression worth catching.
func TestCountdown(t *testing.T) {
	w := httptest.NewRecorder()
	newTestServer(t, "buses").mux.ServeHTTP(w, httptest.NewRequest("GET", "/?id=01019", nil))
	body := w.Body.String()

	arrivals := timeRe.FindAllStringSubmatch(body, -1)
	// 4 real arrivals in the fixture: empty NextBus2/NextBus3 must not render
	if len(arrivals) != 4 {
		t.Fatalf("got %d <time> elements, want 4", len(arrivals))
	}

	for _, a := range arrivals {
		// html.UnescapeString is what the browser's parser does before app.js
		// reads the attribute, turning the &#43; back into a +
		datetime := html.UnescapeString(a[1])
		when, err := time.Parse(time.RFC3339, datetime)
		if err != nil {
			t.Errorf("new Date(%q) would be Invalid Date: %v", datetime, err)
			continue
		}
		// the countdown replaces the text, so it must start as the same time
		if got := html.UnescapeString(strings.TrimSpace(a[2])); got != datetime {
			t.Errorf("<time> text is %q, want %q", got, datetime)
		}
		if when.Location() == time.UTC {
			t.Errorf("%s lost its +08:00 offset, countdown would be 8h out", datetime)
		}
	}

	// app.js looks these up by id; #id is unguarded, so losing it throws
	// before the geolocation redirect and the stations history ever run
	for _, id := range []string{"lastupdated", "stations", "id"} {
		if !hasID(body, id) {
			t.Errorf("app.js needs #%s, not in the render", id)
		}
	}
}

// hasID finds an id attribute, quoted or not, as the template writes both
func hasID(body, id string) bool {
	return regexp.MustCompile(`(?i)\bid=["']?` + id + `["']?[\s>]`).MatchString(body)
}
