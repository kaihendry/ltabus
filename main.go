package main

import (
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
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/apex/gateway/v2"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
	log "golang.org/x/exp/slog"
)

//go:embed static
var static embed.FS

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
		log.Error("failed to create server", err)
	}

	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		err = gateway.ListenAndServe("", server.router)
	} else {
		err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), server.router)
	}
	log.Error("error listening", err)

}

func NewServer(busStopsPath string) (*Server, error) {
	bs, err := loadBusJSON(busStopsPath)
	if err != nil {
		log.Error("unable to load bus stops", err)
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
		log.Error("unable to load static files", err)
	}

	fileServer := http.FileServer(http.FS(directory))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	s.router.HandleFunc("/", s.handleIndex)
	s.router.HandleFunc("/closest", s.handleClosest)
	s.router.HandleFunc("/icon", handleIcon)
	s.router.Use(logger)

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
		log.Error("template failed to parse", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("id")
	var arriving SGBusArrivals

	if id != "" {
		arriving, err = busArrivals(id)
		if err != nil {
			log.Error("failed to retrieve bus timings", err)
			http.Error(w, fmt.Sprintf("datamall API is returning, %s", err.Error()), http.StatusFailedDependency)
			return
		}
	}

	w.Header().Set("X-Version", os.Getenv("version"))

	err = t.ExecuteTemplate(w, "index.html", arriving)
	if err != nil {
		log.Error("template failed to parse", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func busArrivals(stopID string) (arrivals SGBusArrivals, err error) {
	if stopID == "" {
		return arrivals, fmt.Errorf("invalid stop ID")
	}

	url := fmt.Sprintf("http://datamall2.mytransport.sg/ltaodataservice/BusArrivalv2?BusStopCode=%s", stopID)

	client := &http.Client{
		Timeout: 2 * time.Second,
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
		log.Error("failed to decode response", err)
		return
	}

	// Sort by buses arriving first
	sort.Slice(arrivals.Services, func(i, j int) bool {
		return arrivals.Services[i].NextBus.EstimatedArrival < arrivals.Services[j].NextBus.EstimatedArrival
	})

	return
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

type BusStops []BusStop

func loadBusJSON(jsonfile string) (bs BusStops, err error) {
	content, err := static.ReadFile("static/all.json")
	if err != nil {
		log.Error("failed to read file", err)
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
	return template.CSS(fmt.Sprintf("background-color: #%.3x; padding: 0.2em", md5.Sum(data)))
}

func (p Point) distance(p2 Point) float64 {
	latd := p2.lat - p.lat
	lngd := p2.lng - p.lng
	return latd*latd + lngd*lngd
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer slog.Info("served",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			// id value of request
			"stop", r.URL.Query().Get("id"),
			"duration", time.Since(start).Microseconds(),
			// status
			// "requestID", r.Context().Value(requestIDContextKey),
		)

		next.ServeHTTP(w, r)
	})
}
