package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// The JSON reponse representing a city
type city struct {
	CartoID     int64       `json:"cartodb_id"`
	Name        string      `json:"name"`
	Population  int         `json:"population"`
	Coordinates Coordinates `json:"coordinates"`
}

// Setup routes. We only have one but who knows
// what the future holds ;)
func setupAPIRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/id/{cityId}", makeHandlerFunc(getIDHandler)).Methods("GET")

	return r
}

// handlerFunc is a function that handles a request and returns
// a response and a HTTP status code
type handlerFunc func(*http.Request) (interface{}, int)

// makeHandlerFunc will create a new http.HandlerFunc wrapped
// in code that monitors how much time is spent on the request
func makeHandlerFunc(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		// call the actual handler
		res, status := fn(r)
		if status == 0 { // by default, our status is 200 (OK)
			status = http.StatusOK
		}
		// our responses will be JSON objects
		w.Header().Set("Content-Type", "application/json")

		// add the response-time header to the response
		w.Header().Set("API-Response-Time", fmt.Sprintf("%v", time.Since(now)))
		w.WriteHeader(status)
		encoder := json.NewEncoder(w)
		if prettyJSON {
			encoder.SetIndent("", "    ")
		}
		encoder.Encode(res)
	}
}

// getIDHandler handles the requests to /id/<cityid>.
func getIDHandler(r *http.Request) (response interface{}, status int) {
	idS := mux.Vars(r)["cityId"]
	cityID, err := strconv.ParseInt(idS, 10, 64)
	if err != nil { // bad value. CartoDB_IDs should be integers
		return makeError(http.StatusBadRequest, err.Error())
	}

	// find the city
	f, found := featureCollection.FindID(cityID)
	if !found {
		return makeError(http.StatusNotFound, "no city found for CartoDB_ID %d", cityID)
	}

	// if the dist parameter is present, we continue
	// in getCitiesInBox()
	distParam := r.URL.Query()["dist"]
	if len(distParam) == 1 {
		return getCitiesInBox(&f, distParam[0])
	}

	// we're done, we just need to return the city found
	response = map[string]city{"city": featureToCity(f)}
	return
}

// getCitiesInBox handles the case where ?dist=N is included in
// the /id/<id> request. It will compute a bounding box centred
// at the coordinates of f and defined by radius distParam and will
// return a list of cities
func getCitiesInBox(f *Feature, distParam string) (response interface{}, status int) {
	dist, err := strconv.ParseFloat(distParam, 64)
	if err != nil { // bad distance parameter. Should be a float in Km
		return makeError(http.StatusBadRequest, "%q is not a valid distance", distParam)
	}

	// get all nearby features
	feats, err := featureCollection.GetFeaturesNear(f.Properties.CartoDBId, dist)
	if err != nil {
		return makeError(http.StatusInternalServerError, err.Error())
	}

	cities := struct {
		Cities map[string]city `json:"cities"`
	}{
		Cities: make(map[string]city, len(feats)),
	}

	for _, v := range feats {
		cities.Cities[strconv.FormatInt(v.Properties.CartoDBId, 10)] = featureToCity(v)
	}

	return makeOk(cities)
}

// convert a Feature into a city object
func featureToCity(f Feature) city {
	return city{
		CartoID:     f.Properties.CartoDBId,
		Name:        f.Properties.Name,
		Population:  f.Properties.Population,
		Coordinates: f.Geometry.Coordinates,
	}
}

// API response when an error occurs
type apiError struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

// helper to create an apiError object and match the return of handlerFunc
func makeError(status int, format string, a ...interface{}) (interface{}, int) {
	return apiError{Error: fmt.Sprintf(format, a), Status: status}, status
}

// helper to return from handlerFunc with 200 (OK)
func makeOk(a interface{}) (interface{}, int) {
	return a, http.StatusOK
}
