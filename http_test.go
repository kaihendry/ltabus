package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestRoutesAndVisitorCookies(t *testing.T) {
	server, err := NewServer("static/all.json")
	if err != nil {
		t.Fatal(err)
	}
	handler := server.middlewareChain(server.mux)
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(old)
	for _, tc := range []struct {
		path   string
		status int
		page   bool
	}{
		{"/", 200, true}, {"/?id=99999", 200, true},
		{"/nonsense", 404, false}, {"/robots.txt", 404, false},
		{"/static/app.css", 200, false}, {"/icon?stop=99999", 200, false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			logs.Reset()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			want := 0
			if tc.page {
				want = 1
			}
			if cookies := w.Result().Cookies(); len(cookies) != want {
				t.Fatalf("cookies = %v, want %d", cookies, want)
			}
			if strings.Contains(logs.String(), `"unique"`) || !strings.Contains(logs.String(), `"visitor_cookie":false`) {
				t.Fatalf("cookieless request counted as a visitor: %s", logs.String())
			}
			// Any request that sends the cookie back counts, so a first visit
			// is confirmed by the page's own stylesheet, script or icon.
			logs.Reset()
			r := httptest.NewRequest("GET", tc.path, nil)
			r.AddCookie(&http.Cookie{Name: "visitor", Value: "visitor-existing"})
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if len(w.Result().Cookies()) != 0 {
				t.Fatal("returning visitor cookie replaced")
			}
			if !strings.Contains(logs.String(), `"unique":"visitor-existing"`) || !strings.Contains(logs.String(), `"visitor_cookie":true`) {
				t.Fatalf("returned cookie not counted: %s", logs.String())
			}
		})
	}
}

func TestLambdaCookies(t *testing.T) {
	for _, count := range []int{1, 2} {
		handler := lambdaHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &http.Cookie{Name: "visitor", Value: "existing"})
			if count == 2 {
				http.SetCookie(w, &http.Cookie{Name: "second", Value: "value"})
			}
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte{0, 1, 2})
		}))
		response, err := handler(context.Background(), events.APIGatewayV2HTTPRequest{
			RawPath: "/", RequestContext: events.APIGatewayV2HTTPRequestContext{
				HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "GET", Path: "/"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		payload, err := json.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(string(payload)), "set-cookie") {
			t.Fatalf("duplicate cookie headers: %s", payload)
		}
		if len(response.Cookies) != count || response.Cookies[0] != "visitor=existing" {
			t.Fatalf("cookies = %v", response.Cookies)
		}
		if response.StatusCode != 201 || !response.IsBase64Encoded || response.Body != "AAEC" || response.Headers["Content-Type"] != "image/png" {
			t.Fatalf("response changed: %s", payload)
		}
	}
}
