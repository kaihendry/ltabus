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

// testNow freezes the clock, so the arrivals stop 99999 makes up - and the
// golden render of them - are the same on every run.
var testNow = time.Date(2026, 8, 24, 21, 0, 0, 0, time.FixedZone("SGT", 8*60*60))

// newTestServer serves the test stop from a fixed clock. Pass down to make
// the datamall lookup fail, which is the only case that needs it: the test
// stop never reaches datamall.
func newTestServer(t *testing.T, down bool) *Server {
	t.Helper()
	s, err := NewServer("static/all.json")
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s.now = func() time.Time { return testNow }
	if down {
		s.arrivals = func(string) (SGBusArrivals, error) {
			return SGBusArrivals{}, os.ErrDeadlineExceeded
		}
	}
	return s
}

// TestRender renders the whole page against fixtures and diffs it against a
// golden file. Run `go test -update` after intentional HTML changes and read
// the git diff to review what moved.
func TestRender(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		down   bool
		status int
	}{
		{name: "index", url: "/", status: http.StatusOK},
		{name: "buses", url: "/?id=" + testStopCode, status: http.StatusOK},
		{name: "datamalldown", url: "/?id=01019", down: true, status: http.StatusFailedDependency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			newTestServer(t, tt.down).mux.ServeHTTP(w, httptest.NewRequest("GET", tt.url, nil))

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
	newTestServer(t, false).mux.ServeHTTP(w, httptest.NewRequest("GET", "/?id="+testStopCode, nil))
	body := w.Body.String()

	for _, want := range []string{
		"<title>Singapore bus arrival times for stop 99999</title>",
		testStopName,                      // nameBusStop func is wired up
		`href='/icon?stop=99999'`,         // favicon per stop
		`content="20"`,                    // page refreshes itself
		`name="robots" content="noindex"`, // the test stop stays out of search
		`class="load-seats"`,              // loadClass SEA
		`class="load-standing"`,           // loadClass SDA
		`class="load-full"`,               // loadClass LSD
		// first arrival is testNow + 1m30s, + escaped by html/template
		`dateTime="2026-08-24T21:01:30&#43;08:00"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("render is missing %s", want)
		}
	}

	// soonest bus first: testArrivals lists 666 before 42
	if i, j := strings.Index(body, ">42<"), strings.Index(body, ">666<"); i == -1 || i > j {
		t.Errorf("services not sorted by arrival: 42 at %d, 666 at %d", i, j)
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
	newTestServer(t, false).mux.ServeHTTP(w, httptest.NewRequest("GET", "/?id="+testStopCode, nil))
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
