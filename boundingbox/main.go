package main

import (
	"flag"
	"fmt"
)

func main() {
	lat := flag.Float64("lat", 46.716993, "latitude")
	lon := flag.Float64("lon", -71.269204, "longitude")
	dist := flag.Float64("dist", 2, "distance (radius) in Km")

	flag.Usage = func() {
		fmt.Printf("Usage: boundingbox [options]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	box := GetBoundingBox(Coordinate{*lat, *lon}, *dist)

	fmt.Println(box)
}
