package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var featureCollection *FeatureCollection

var (
	dataFile      string
	listen        string
	ignoreZeroPop bool // ignore features with 0 population
	prettyJSON    bool // whether to indent JSON responses
)

func init() {
	flag.StringVar(&dataFile, "filename", "canada_cities.geojson", "A geojson file containing the data")
	flag.StringVar(&dataFile, "f", "canada_cities.geojson", "A geojson file containing the data (shorthand)")
	flag.StringVar(&listen, "l", ":8000", "Where the server will listen to")
	flag.BoolVar(&ignoreZeroPop, "nz", false, "Ignore features with population 0")
	flag.BoolVar(&prettyJSON, "pretty", false, "Indent JSON responses")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.Printf("Reading data from %q\n", dataFile)
	var err error
	featureCollection, err = NewFeatureCollectionFromFile(dataFile)
	if err != nil {
		log.Fatalln("could create the feature collection: ", err)
	}

	log.Printf("Read %d features (%d indexed.) All ready.\n", len(featureCollection.Features), featureCollection.Indexed())
	log.Fatal(http.ListenAndServe(listen, setupAPIRouter()))
}

func usage() {
	fmt.Println("Usage: geotest [options]")
	flag.PrintDefaults()
}
