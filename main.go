package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"strconv"
	"time"

	"html/template"

	"github.com/apex/log"
	jsonloghandler "github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/gorilla/mux"
)

type key int

const (
	logger key = iota
	visitor
)

// NextBus describes when the bus is coming
type NextBus struct {
	OriginCode       string `json:"OriginCode"`
	DestinationCode  string `json:"DestinationCode"`
	EstimatedArrival string `json:"EstimatedArrival"`
	Latitude         string `json:"Latitude"`
	Longitude        string `json:"Longitude"`
	VisitNumber      string `json:"VisitNumber"`
	Load             string `json:"Load"`
	Feature          string `json:"Feature"`
	Type             string `json:"Type"`
}

// SGBusArrivals describes the response from the datamall API
type SGBusArrivals struct {
	OdataMetadata string `json:"odata.metadata"`
	BusStopCode   string `json:"BusStopCode"`
	Services      []struct {
		ServiceNo string  `json:"ServiceNo"`
		Operator  string  `json:"Operator"`
		NextBus   NextBus `json:"NextBus"`
		NextBus2  NextBus `json:"NextBus2"`
		NextBus3  NextBus `json:"NextBus3"`
	} `json:"Services"`
}

var bs BusStops

func init() {
	if os.Getenv("UP_STAGE") != "" {
		log.SetHandler(jsonloghandler.Default)
	} else {
		log.SetHandler(text.Default)
	}
}

func main() {

	bs, _ = loadBusJSON("all.json")
	log.Infof("Loaded %d bus stops", len(bs))

	app := mux.NewRouter()
	app.HandleFunc("/", handleIndex)
	app.HandleFunc("/closest", handleClosest)
	app.HandleFunc("/icon", handleIcon)
	app.Use(addContextMiddleware)

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func handleClosest(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	closest := bs.closest(Point{lat: lat, lng: lng})
	http.Redirect(w, r, fmt.Sprintf("/?id=%s", closest.BusStopCode), 302)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("UP_STAGE") != "production" {
		w.Header().Set("X-Robots-Tag", "none")
	}

	log, ok := r.Context().Value(logger).(*log.Entry)
	if !ok {
		http.Error(w, "Unable to get logging context", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"nameBusStopID": func(s string) string { return bs.nameBusStopID(s) },
		"totalstops":    func() int { return len(bs) },
		"getenv":        os.Getenv,
	}

	t, err := template.New("").Funcs(funcs).ParseFiles("templates/index.html")
	if err != nil {
		log.WithError(err).Error("template failed to parse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("id")
	var arriving SGBusArrivals

	if id != "" {
		arriving, err = busArrivals(id)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bus timings")
		}
		log.WithField("input", id).Info("serving")
	}

	err = t.ExecuteTemplate(w, "index.html", arriving)
	if err != nil {
		log.WithError(err).Error("template failed to execute")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func busArrivals(stopID string) (arrivals SGBusArrivals, err error) {

	if stopID == "" {
		return
	}

	ctx := log.WithFields(
		log.Fields{
			"stopID": stopID,
		})

	url := fmt.Sprintf("https://api.mytransport.sg/ltaodataservice/BusArrivalv2/?BusStopCode=%s", stopID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("AccountKey", os.Getenv("accountkey"))

	var t0, t1, t2, t3, t4, t5, t6 time.Time

	trace := &httptrace.ClientTrace{
		DNSStart:             func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		ConnectStart:         func(_, _ string) { t1 = time.Now() },
		ConnectDone:          func(_, _ string, _ error) { t2 = time.Now() },
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	t7 := time.Now() // after read body

	// https://github.com/davecheney/httpstat/blob/master/main.go#L343
	ctx.WithFields(
		log.Fields{
			"dnslookup":         t1.Sub(t0),
			"tcp connection":    t2.Sub(t1),
			"tls handshake":     t6.Sub(t5),
			"server processing": t4.Sub(t3),
			"content transfer":  t7.Sub(t4),
			"namelookup":        t1.Sub(t0),
			"connect":           t2.Sub(t0),
			"pretransfer":       t3.Sub(t0),
			"starttransfer":     t4.Sub(t0),
			"total":             t7.Sub(t0),
			"status":            res.StatusCode,
		}).Info("LTA API")

	if res.StatusCode != http.StatusOK {
		return arrivals, fmt.Errorf("Bad response: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&arrivals)
	if err != nil {
		log.WithError(err).Error("failed to decode response")
		return
	}

	// Sort by buses arriving first
	sort.Slice(arrivals.Services, func(i, j int) bool {
		return arrivals.Services[i].NextBus.EstimatedArrival < arrivals.Services[j].NextBus.EstimatedArrival
	})

	return
}

func addContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("visitor")
		logging := log.WithFields(
			log.Fields{
				"id":      r.Header.Get("X-Request-Id"),
				"country": r.Header.Get("Cloudfront-Viewer-Country"),
				"ua":      r.UserAgent(),
			})
		if cookie != nil {
			cvisitor := context.WithValue(r.Context(), visitor, cookie.Value)
			logging = logging.WithField("visitor", cookie.Value)
			clog := context.WithValue(cvisitor, logger, logging)
			next.ServeHTTP(w, r.WithContext(clog))
		} else {
			visitorID, _ := generateRandomString(24)
			// log.Infof("Generating vistor id: %s", visitorID)
			expiration := time.Now().Add(365 * 24 * time.Hour)
			setCookie := http.Cookie{Name: "visitor", Value: visitorID, Expires: expiration}
			http.SetCookie(w, &setCookie)
			cvisitor := context.WithValue(r.Context(), visitor, visitorID)
			logging = logging.WithField("visitor", visitorID)
			clog := context.WithValue(cvisitor, logger, logging)
			next.ServeHTTP(w, r.WithContext(clog))
		}
	})
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
