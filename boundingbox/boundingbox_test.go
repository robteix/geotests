package main

import (
	"fmt"
	"testing"
)

func TestGetBoundingBox(t *testing.T) {

	type TestCoord struct {
		Lon      float64
		Lat      float64
		Distance float64
		Want     string
	}

	testCoords := []TestCoord{
		TestCoord{-71.269204, 46.716993, 2, "46.699007,-71.295438,46.734979,-71.242970"},
		TestCoord{-0, 0, 1000, "-8.993202,-8.993202,8.993202,8.993202"},
		TestCoord{147.3494, 64.7511, 1000, "55.757898,125.851468,73.744302,168.847332"}, // near north pole
		TestCoord{180, 90, 1000, "81.006798,-180.000000,90.000000,180.000000"},          // north pole
		TestCoord{78, 180, 5000, "135.033990,-180.000000,90.000000,180.000000"},         // Ross Dependency, NZ (near south pole)
	}

	for _, tc := range testCoords {
		box := GetBoundingBox(Coordinate{tc.Lat, tc.Lon}, tc.Distance)
		res := fmt.Sprintf("%v", box)
		if res != tc.Want {
			t.Errorf("got %s, want %s", res, tc.Want)
		}
	}
}
