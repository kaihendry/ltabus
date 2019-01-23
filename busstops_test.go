package main

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func init() {
	var err error
	bs, err = loadBusJSON("all.json")
	fmt.Println("Number of bus stops", len(bs))
	if err != nil {
		log.Fatal(err)
	}
}

func Benchmark_closest(t *testing.B) {
	type args struct {
		location Point
	}
	tests := []struct {
		name     string
		BusStops BusStops
		args     args
		wantB    BusStop
	}{
		{
			name:     "Middle Earth",
			BusStops: bs,
			args: args{
				location: Point{
					lat: 0.0,
					lng: 0.0,
				},
			},
			wantB: BusStop{
				BusStopCode: "25751",
				RoadName:    "Tuas Sth Ave 5",
				Description: "BEF TUAS STH AVE 14",
				Latitude:    1.27637,
				Longitude:   103.621508,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.B) {
			if gotB := tt.BusStops.closest(tt.args.location); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("BusStops.closest() = %+v, want %+v", gotB, tt.wantB)
			}
		})
	}
}

func TestBusStops_closest(t *testing.T) {
	type args struct {
		location Point
	}
	tests := []struct {
		name     string
		BusStops BusStops
		args     args
		wantB    BusStop
	}{
		{
			name:     "Middle Earth",
			BusStops: bs,
			args: args{
				location: Point{
					lat: 0.0,
					lng: 0.0,
				},
			},
			wantB: BusStop{
				BusStopCode: "25751",
				RoadName:    "Tuas Sth Ave 5",
				Description: "BEF TUAS STH AVE 14",
				Latitude:    1.27637,
				Longitude:   103.621508,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := tt.BusStops.closest(tt.args.location); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("BusStops.closest() = %+v, want %+v", gotB, tt.wantB)
			}
		})
	}
}

func TestBusStops_nameBusStopID(t *testing.T) {
	type args struct {
		busid string
	}
	tests := []struct {
		name            string
		BusStops        BusStops
		args            args
		wantDescription string
	}{
		{
			name:     "Bras Basah",
			BusStops: bs,
			args: args{
				busid: "01019",
			},
			wantDescription: "Bras Basah Cplx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDescription := tt.BusStops.nameBusStopID(tt.args.busid); gotDescription != tt.wantDescription {
				t.Errorf("BusStops.nameBusStopID() = %v, want %v", gotDescription, tt.wantDescription)
			}
		})
	}
}
