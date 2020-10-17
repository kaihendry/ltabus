package main

import (
	"reflect"
	"testing"
)

// We need to ensure compiler actually returns the value
// A clever compiler might optimise it out, rendering our
// benchmarking results incorrect

var stop BusStop

func Benchmark_closest(b *testing.B) {
	bs, _ := loadBusJSON("all.json")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		stop = bs.closest(Point{}) // ensure the compiler returns a value
	}
}

func TestBusStops_closest(t *testing.T) {
	bs, _ := loadBusJSON("all.json")
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
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			if gotB := tt.BusStops.closest(tt.args.location); !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("BusStops.closest() = %+v, want %+v", gotB, tt.wantB)
			}
		})
	}
}

func TestBusStops_nameBusStopID(t *testing.T) {
	bs, _ := loadBusJSON("all.json")
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
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			if gotDescription := tt.BusStops.nameBusStop(tt.args.busid); gotDescription != tt.wantDescription {
				t.Errorf("BusStops.nameBusStopID() = %v, want %v", gotDescription, tt.wantDescription)
			}
		})
	}
}
