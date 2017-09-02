package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
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
		box, _ := GetBoundingBox(Coordinates{tc.Lon, tc.Lat}, tc.Distance)
		res := fmt.Sprintf("%v", box)
		if res != tc.Want {
			t.Errorf("got %s, want %s", res, tc.Want)
		}
	}
}

func TestIndexItems(t *testing.T) {
	fc := FeatureCollection{}
	// there are three features. All share the same point
	// and two share the same cartodb.
	fc.Features = []Feature{
		Feature{
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: Coordinates{0, 0},
			},
			Properties: FeatureProperties{
				CartoDBId: 0,
				Name:      "city0",
			},
		},
		Feature{
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: Coordinates{0, 0},
			},
			Properties: FeatureProperties{
				CartoDBId: 1,
				Name:      "city1",
			},
		},
		Feature{
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: Coordinates{0, 0},
			},
			Properties: FeatureProperties{
				CartoDBId: 0,
				Name:      "city2",
			},
		},
	}

	fc.indexItems()

	// the lat index doesn't care about duplicates because
	// it indexes only the latitude and the position on the
	// original slice. So we'll have all three on the index
	if fc.latIndex.Len() != 3 {
		t.Errorf("got %d indexes, want 2", fc.latIndex.Len())
	}

	// the ID index will not hold two items with the same ID, so
	// here we have to have only 2 instead of 3 items indexed.
	if fc.idIndex.Len() != 2 {
		t.Errorf("got %d items, want 2", fc.idIndex.Len())
	}
}

func TestFindId(t *testing.T) {
	const nfeats = 1000
	fc := FeatureCollection{
		Features: make([]Feature, nfeats),
	}
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	ids := make([]int, nfeats)
	for i := 0; i < nfeats; i++ {
		id := r.Intn(30000)
		ids[i] = id
		fc.Features[i] = Feature{
			Properties: FeatureProperties{
				CartoDBId: id,
				Name:      fmt.Sprintf("city%d", i),
			},
		}
	}

	fc.indexItems()

	for _, id := range ids {
		c, found := fc.FindID(id)
		if !found {
			t.Errorf("could not find previously added city %d", id)
		}
		if c.Properties.CartoDBId != id {
			t.Errorf("got id %d, want %d", c.Properties.CartoDBId, id)
		}
	}

}
