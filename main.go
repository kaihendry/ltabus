package main

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/apex/gateway/v2"
	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/gorilla/mux"
)

type key int

//go:embed static
var static embed.FS

var Version string

const (
	logger key = iota
	visitor
)

type tracingRoundTripper struct {
	next http.RoundTripper
	dest *log.Logger
}

func (rt *tracingRoundTripper) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	var (
		start     = time.Now()
		dnsStart  time.Duration
		firstByte time.Duration
	)

	defer func() {
		f := log.Fields{
			"dns_start_ms":  dnsStart.Milliseconds(),
			"first_byte_ms": firstByte.Milliseconds(),
			"total_ms":      time.Since(start).Milliseconds(),
			"url":           r.URL.String(),
		}
		switch {
		case err == nil:
			f["response_code"] = resp.StatusCode
		case err != nil:
			f["error"] = err.Error()
		}
		rt.dest.WithFields(f).Trace("fetch")
	}()

	tr := &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { dnsStart = time.Since(start) },
		GotFirstResponseByte: func() { firstByte = time.Since(start) },
	}

	ctx := httptrace.WithClientTrace(r.Context(), tr)

	return rt.next.RoundTrip(r.WithContext(ctx))
}

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

type Server struct {
	router   *mux.Router
	busStops BusStops
}

func main() {

	server, err := NewServer("all.json")
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	log.SetHandler(jsonhandler.Default)
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		err = gateway.ListenAndServe("", server.router)
	} else {
		log.SetHandler(text.Default)
		err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), server.router)
	}
	log.WithError(err).Fatal("error listening")

}

func NewServer(busStopsPath string) (*Server, error) {

	bs, err := loadBusJSON(busStopsPath)
	if err != nil {
		log.WithError(err).Fatal("unable to load bus stops")
	}

	srv := Server{
		router:   mux.NewRouter(),
		busStops: bs,
	}

	srv.routes()

	return &srv, nil

}

func (s *Server) routes() {

	directory, err := fs.Sub(static, "static")
	if err != nil {
		log.WithError(err).Fatal("unable to load static files")
	}
	fileServer := http.FileServer(http.FS(directory))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	s.router.HandleFunc("/", s.handleIndex)
	s.router.HandleFunc("/closest", s.handleClosest)
	s.router.HandleFunc("/icon", handleIcon)

	s.router.Use(addContextMiddleware)
}

func (s *Server) handleClosest(w http.ResponseWriter, r *http.Request) {
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

	closest := s.busStops.closest(Point{lat: lat, lng: lng})
	http.Redirect(w, r, fmt.Sprintf("/?id=%s", closest.BusStopCode), http.StatusFound)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	log, ok := r.Context().Value(logger).(*log.Entry)
	if !ok {
		http.Error(w, "Unable to get logging context", http.StatusInternalServerError)
		return
	}

	funcs := template.FuncMap{
		"totalStops":   func() int { return len(s.busStops) },
		"nameBusStop":  func(id string) string { return s.busStops.nameBusStop(id) },
		"styleBusStop": func(id string) template.CSS { return styleBusStop(id) },
		"getEnv":       os.Getenv,
	}

	// set html content type
	w.Header().Set("Content-Type", "text/html")

	t, err := template.New("").Funcs(funcs).ParseFS(static, "static/index.html")
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
			http.Error(w, fmt.Sprintf("datamall API is returning, %s", err.Error()), http.StatusFailedDependency)
			return
		}
		log.WithField("input", id).Info("serving")
	}

	w.Header().Set("X-Version", Version)
	err = t.ExecuteTemplate(w, "index.html", arriving)
	if err != nil {
		log.WithError(err).Error("template failed to parse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func busArrivals(stopID string) (arrivals SGBusArrivals, err error) {

	if stopID == "" {
		return arrivals, fmt.Errorf("invalid stop ID")
	}

	ctx := log.WithFields(
		log.Fields{
			"stopID": stopID,
		})

	url := fmt.Sprintf("http://datamall2.mytransport.sg/ltaodataservice/BusArrivalv2?BusStopCode=%s", stopID)

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &tracingRoundTripper{
			next: http.DefaultTransport,
			dest: ctx.Logger,
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// get accountkey from env
	req.Header.Add("AccountKey", os.Getenv("accountkey"))

	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return arrivals, fmt.Errorf("bad response: %d", res.StatusCode)
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

func ms(d time.Duration) int {
	return int(d / time.Millisecond)
}

// Point is a geo co-ordinate
type Point struct {
	lat float64
	lng float64
}

// BusStop describes a Singaporean (LTA) bus stop
type BusStop struct {
	BusStopCode string  `json:"BusStopCode"`
	RoadName    string  `json:"RoadName"`
	Description string  `json:"Description"`
	Latitude    float64 `json:"Latitude"`
	Longitude   float64 `json:"Longitude"`
}

// BusStops are many bus stops
type BusStops []BusStop

func loadBusJSON(jsonfile string) (bs BusStops, err error) {
	content, err := static.ReadFile("static/all.json")
	if err != nil {
		log.Error("failed to read file")
		return
	}
	err = json.Unmarshal(content, &bs)
	if err != nil {
		return
	}

	return
}

func (bs BusStops) closest(location Point) BusStop {
	c := -1
	closestSoFar := math.Inf(1)
	for i := range bs {
		distance := location.distance(Point{bs[i].Latitude, bs[i].Longitude})
		if distance < closestSoFar {
			// Set the return
			c = i
			// Record closest distance
			closestSoFar = distance
		}
	}
	return bs[c]
}

func (bs BusStops) nameBusStop(busStopID string) (description string) {
	for _, p := range bs {
		if busStopID == p.BusStopCode {
			return p.Description
		}
	}
	return ""
}

func styleBusStop(busStopID string) (style template.CSS) {
	data := []byte(busStopID)
	log.Infof("underline color #%.3x", md5.Sum(data))
	return template.CSS(fmt.Sprintf("background-color: #%.3x; padding: 0.2em", md5.Sum(data)))
}

// distance calculates the distance between two points
func (p Point) distance(p2 Point) float64 {
	latd := p2.lat - p.lat
	lngd := p2.lng - p.lng
	return latd*latd + lngd*lngd
}
