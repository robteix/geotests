package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/google/btree"
)

// FeatureCollection represents a collection of Feature objects
type FeatureCollection struct {
	Features []Feature // A slice of Feature objects

	idIndex  *btree.BTree
	latIndex *btree.BTree
}

// Feature represents a point such as a city
type Feature struct {
	Geometry   Geometry          `json:"geometry"`
	Properties FeatureProperties `json:"properties"`
}

// Geometry is a GeoJSON object defining a point (in our current test; it can
// define other types of Geometry elsewhere)
type Geometry struct {
	Type        string      `json:"Type"`        // The type of geometry object (always "Point" in our test)
	Coordinates Coordinates `json:"coordinates"` // The actuall Coordinates defining the point
}

// Coordinates contains coordinates of a point. In order, they are
// longiture, latitude, and altitude. Longitude and latitude are required,
// all others are optional
type Coordinates []float64

// Lat returns the latitude part of the coordinates
func (c Coordinates) Lat() float64 {
	if len(c) < 2 {
		return 0
	}
	return c[1]
}

// Lon returns the longitude part of the coordinates
func (c Coordinates) Lon() float64 {
	if len(c) < 2 {
		return 0
	}
	return c[0]
}

// FeatureProperties contains several properties of a Feature
type FeatureProperties struct {
	Name       string    `json:"name"`
	PlaceKey   string    `json:"place_key"`
	Capital    string    `json:"capital"`
	Population int       `json:"population"`
	PClass     string    `json:"pclass"`
	CartoDBId  int64     `json:"cartodb_id"`
	Createdat  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewFeatureCollectionFromFile create and populates a new FeatureCollection
// from a file of GeoJSON data
func NewFeatureCollectionFromFile(filename string) (*FeatureCollection, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	featCollection := new(FeatureCollection)

	err = json.Unmarshal(data, featCollection)
	if err != nil {
		return nil, err
	}

	// create indexes
	featCollection.idIndex = btree.New(2)
	featCollection.latIndex = btree.New(2)
	for k, v := range featCollection.Features {
		if ignoreZeroPop && v.Properties.Population == 0 {
			continue
		}
		latItem := latIndex{
			Index:    k,
			Latitude: v.Geometry.Coordinates[1],
		}
		idItem := idIndex{
			Index:   k,
			CartoID: v.Properties.CartoDBId,
		}
		featCollection.idIndex.ReplaceOrInsert(idItem)
		featCollection.latIndex.ReplaceOrInsert(latItem)
	}

	return featCollection, nil
}

// FindID tries to find and fetch a Feature from its cartoid
func (fc FeatureCollection) FindID(code int64) (feature Feature, found bool) {
	i := fc.idIndex.Get(idIndex{CartoID: code})
	if i == nil {
		return Feature{}, false
	}

	itemIdx := i.(idIndex)
	return fc.Features[itemIdx.Index], true
}

// GetFeaturesNear returns a list of all features found
// within distance in Km of the FEATURES identified by its cartoId
func (fc FeatureCollection) GetFeaturesNear(cartoID int64, distance float64) ([]Feature, error) {
	origin, found := fc.FindID(cartoID)
	if !found {
		return nil, fmt.Errorf("could not find a city with a CartoDBId %d", cartoID)
	}
	box, err := GetBoundingBox(origin.Geometry.Coordinates, distance)
	if err != nil {
		return nil, err
	}

	inLat := make([]Feature, 0)
	fc.latIndex.DescendRange(latIndex{Latitude: box.Max.Lat()}, latIndex{Latitude: box.Min.Lat()}, func(i btree.Item) bool {
		idx := i.(latIndex).Index
		feat := fc.Features[idx]
		if feat.Properties.CartoDBId == cartoID && excludeOrigin {
			return true
		}
		featLon := feat.Geometry.Coordinates[0]
		if featLon >= box.Min.Lon() && featLon <= box.Max.Lon() {
			inLat = append(inLat, feat)
		}
		return true

	})

	return inLat, nil
}

func (fc FeatureCollection) Indexed() int {
	return fc.latIndex.Len()
}

// The latitude-based index that implements the btree.Item
// interface and will be stored in a btree for quick retrieval
type latIndex struct {
	Index    int
	Latitude float64
}

// Implements btree.Item and tests whether item a is less than item b
// based on its latitude
func (a latIndex) Less(b btree.Item) bool {
	li := b.(latIndex)
	if a.Latitude == li.Latitude {
		// differenciate two cities with the exact same
		// latitude by Index)
		return a.Index < li.Index
	}
	return a.Latitude < li.Latitude
}

// The cartoid-based index that implements a btree.Item
// interface and will be stored in a btree for quick retrieval
type idIndex struct {
	Index   int
	CartoID int64
}

// Implements btree.Item and tests whether item a is less than item b
// based on the cartoid
func (a idIndex) Less(b btree.Item) bool {
	return a.CartoID < b.(idIndex).CartoID
}
