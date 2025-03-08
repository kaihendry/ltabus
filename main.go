package main

import (
	"crypto/md5"
	"embed"
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"log/slog"

	"github.com/apex/gateway/v2"
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
	router   *chi.Mux
	busStops BusStops
}

func getLogger(logLevel string) *slog.Logger {
	levelVar := slog.LevelVar{}

	if logLevel != "" {
		if err := levelVar.UnmarshalText([]byte(logLevel)); err != nil {
			panic(fmt.Sprintf("Invalid log level %s: %v", logLevel, err))
		}
	}
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: levelVar.Level(),
		}))
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: levelVar.Level(),
	}))
}

func main() {
	server, err := NewServer("static/all.json")
	if err != nil {
		slog.Error("failed to create server", "error", err)
	}

	slog.SetDefault(getLogger(os.Getenv("LOGLEVEL")))

	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		slog.Info("starting server", "version", os.Getenv("VERSION"))
		err = gateway.ListenAndServe("", server.router)
	} else {
		slog.Info("starting local server", "version", os.Getenv("VERSION"), "port", os.Getenv("PORT"))
		err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), server.router)
	}

	slog.Error("error listening", "error", err)
}

func NewServer(busStopsPath string) (*Server, error) {
	bs, err := loadBusJSON(busStopsPath)
	if err != nil {
		slog.Error("unable to load bus stops", "error", err)
	}

	srv := Server{
		router:   chi.NewRouter(),
		busStops: bs,
	}

	srv.router.Use(middleware.RequestID)
	srv.router.Use(uniqueVisitor)
	srv.router.Use(logRequest)
	srv.router.Use(middleware.Recoverer)

	srv.router.Get("/", srv.handleIndex)
	srv.router.Get("/closest", srv.handleClosest)
	srv.router.Get("/icon", handleIcon)

	directory, err := fs.Sub(static, "static")
	if err != nil {
		slog.Error("unable to load static files", "error", err)
	}
	fileServer := http.FileServer(http.FS(directory))
	//srv.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	srv.router.Mount("/static", http.StripPrefix("/static", fileServer))

	return &srv, nil
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
		slog.Error("template failed to parse", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("id")
	var arriving SGBusArrivals

	if id != "" {
		arriving, err = busArrivals(id)
		if err != nil {
			slog.Error("failed to retrieve bus timings", "error", err)
			http.Error(w, fmt.Sprintf("datamall API is returning, %s", err.Error()), http.StatusFailedDependency)
			return
		}
	}

	w.Header().Set("X-Version", os.Getenv("VERSION"))

	err = t.ExecuteTemplate(w, "index.html", arriving)
	if err != nil {
		slog.Error("template failed to parse", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func busArrivals(stopID string) (arrivals SGBusArrivals, err error) {
	if stopID == "" {
		return arrivals, fmt.Errorf("invalid stop ID")
	}

	url := fmt.Sprintf("https://datamall2.mytransport.sg/ltaodataservice/v3/BusArrival?BusStopCode=%s", stopID)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("AccountKey", os.Getenv("ACCOUNTKEY"))

	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return arrivals, fmt.Errorf("bad response: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&arrivals)
	if err != nil {
		slog.Error("failed to decode response", "error", err)
		return
	}

	// Sort by buses arriving first
	sort.Slice(arrivals.Services, func(i, j int) bool {
		return arrivals.Services[i].NextBus.EstimatedArrival < arrivals.Services[j].NextBus.EstimatedArrival
	})

	return
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
	content, err := static.ReadFile(jsonfile)
	if err != nil {
		slog.Error("failed to read file", "error", err)
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

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			// Log response headers here
			for name, values := range ww.Header() {
				for _, value := range values {
					slog.Debug("response header", "name", name, "value", value)
				}
			}

			slog.Info("response",
				"req_method", r.Method,
				"req_ip", r.RemoteAddr,
				"req_path", r.RequestURI,
				"res_status", ww.Status(),
				"res_size", ww.BytesWritten(),
				"res_content_type", ww.Header().Get("Content-Type"),
				"duration", time.Since(start).Milliseconds(),
			)
		}()
		next.ServeHTTP(ww, r)
	})
}

func uniqueVisitor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("visitor")
		// logging := log.WithFields(
		// 	log.Fields{
		// 		"id":      r.Header.Get("X-Request-Id"),
		// 		"country": r.Header.Get("Cloudfront-Viewer-Country"),
		// 		"ua":      r.UserAgent(),
		// 	})
		if cookie != nil {
			slog.Info("return visitor", "unique", cookie.Value)
		} else {
			setCookie := http.Cookie{Name: "visitor", Value: fmt.Sprint("visitor-", time.Now().UnixMilli()), Expires: time.Now().Add(365 * 24 * time.Hour)}
			http.SetCookie(w, &setCookie)
		}
		next.ServeHTTP(w, r)

	})
}
