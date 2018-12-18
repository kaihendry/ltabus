package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"

	"html/template"

	"github.com/apex/log"
	"github.com/gorilla/mux"
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

// SGBusArrivals describes the response of from the datamall API
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

func main() {

	bs, _ = loadBusJSON("all.json")

	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()

	app.HandleFunc("/", handleIndex).Methods("GET")
	app.HandleFunc("/closest", handleClosest).Methods("GET")
	app.HandleFunc("/icon", handleIcon).Methods("GET")

	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func handleClosest(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	closest := bs.closest(Point{lat: lat, lng: lng})
	// fmt.Fprintf(w, "%#v", closest)
	http.Redirect(w, r, fmt.Sprintf("/?id=%s", closest.BusStopCode), 302)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("UP_STAGE") != "production" {
		w.Header().Set("X-Robots-Tag", "none")
	}

	funcs := template.FuncMap{"nameBusStopID": func(s string) string { return bs.nameBusStopID(s) }}
	// funcs := template.FuncMap{"nameBusStopID": bs.nameBusStopID }}

	t, err := template.New("").Funcs(funcs).ParseFiles("templates/index.html")

	if err != nil {
		log.WithError(err).Error("template failed to parse")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	arriving, err := busArrivals(r.URL.Query().Get("id"))
	if err != nil {
		log.WithError(err).Error("failed to retrieve bus timings")
	}

	// log.Infof("%+v", arriving)
	t.ExecuteTemplate(w, "index.html", arriving)

}

func busArrivals(id string) (arrivals SGBusArrivals, err error) {

	log.Infof("Looking up %s", id)
	url := fmt.Sprintf("http://datamall2.mytransport.sg/ltaodataservice/BusArrivalv2/?BusStopCode=%s", id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("AccountKey", os.Getenv("accountkey"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

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
