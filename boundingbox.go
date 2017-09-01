package main

import (
	"errors"
	"fmt"
	"math"
)

// approximate radius of the Earth
const earthRadius = 6371.01

// degreeToRadian converts a value specified in degrees to radians
func degreeToRadian(degree float64) float64 {
	return degree * math.Pi / 180
}

// radianToDegree converts a value in radians to degrees
func radianToDegree(radian float64) float64 {
	return 180 * radian / math.Pi
}

// BoundingBox is a box defined by two coordinates
// in a sphere
type BoundingBox struct {
	Min Coordinates
	Max Coordinates
}

func (bb BoundingBox) String() string {
	return fmt.Sprintf("%.6f,%.6f,%.6f,%.6f", bb.Min.Lat(), bb.Min.Lon(), bb.Max.Lat(), bb.Max.Lon())
}

// These constants define the boundaries to test if and of the
// poles or the 180th meridian are within the calculated box
const (
	southPole = -1 * math.Pi / 2 // Latitude of the South Pole
	northPole = math.Pi / 2      // Latidure of the North Pole
	min180th  = math.Pi * -1     // Longitude "west" of the 180th meridian
	max180th  = math.Pi          // Longitude "east" of the 180th meridian
)

// GetDistance calculates the distance in kilometres between points a and b
func GetDistance(a Coordinates, b Coordinates) float64 {
	// convert from degree to rad
	rLatA := degreeToRadian(a.Lat())
	rLonA := degreeToRadian(a.Lon())
	rLatB := degreeToRadian(b.Lat())
	rLonB := degreeToRadian(b.Lon())

	return math.Acos(math.Sin(rLatA)*math.Sin(rLatB)+math.Cos(rLatA)*math.Cos(rLatB)*math.Cos(rLonA-rLonB)) * earthRadius
}

// GetBoundingBox returns the minimun box that encloses a circle
// centred at centre and with a radius of distance kilometres
func GetBoundingBox(centre Coordinates, distance float64) (BoundingBox, error) {
	if centre == nil || len(centre) < 2 {
		return BoundingBox{}, errors.New("invalid centre coordinates")
	}

	angularDistance := distance / earthRadius

	// convert to radians
	rLat := degreeToRadian(centre.Lat())
	rLon := degreeToRadian(centre.Lon())

	latmin := rLat - angularDistance
	latmax := rLat + angularDistance

	if latmin <= southPole || latmax > northPole {
		// one of the poles is within the boundaries
		return BoundingBox{
			Min: Coordinates{
				radianToDegree(min180th),
				radianToDegree(math.Max(latmin, southPole)),
			},
			Max: Coordinates{
				radianToDegree(max180th),
				radianToDegree(math.Min(latmax, northPole)),
			},
		}, nil
	}

	// none of the poles is there, so we do:
	//     lonmin = lonT1 = lon - Δlon
	//     lonmax = lonT2 = lon + Δlon
	// where
	//     Δlon = arcsin(sin(r)/cos(lat))
	Δlon := math.Asin(math.Sin(angularDistance) / math.Cos(rLat))
	lonmin := rLon - Δlon
	lonmax := rLon + Δlon

	// check if the 180th is anywhere within the boundaries
	if lonmin < min180th {
		lonmin += 2 * math.Pi
	}
	if lonmax > max180th {
		lonmax -= 2 * math.Pi
	}

	return BoundingBox{
		Min: Coordinates{
			radianToDegree(lonmin),
			radianToDegree(latmin),
		},
		Max: Coordinates{
			radianToDegree(lonmax),
			radianToDegree(latmax),
		},
	}, nil
}
