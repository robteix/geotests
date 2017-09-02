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

type handlerFunc func(*http.Request) (interface{}, int)

func makeHandlerFunc(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		res, status := fn(r)
		w.Header().Set("API-Response-Time", fmt.Sprintf("%v", time.Since(now)))
		w.Header().Set("Content-Type", "application/json")
		if status == 0 {
			status = http.StatusOK
		}
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
	if err != nil {
		return makeError(http.StatusBadRequest, err.Error())
	}

	f, found := featureCollection.FindID(cityID)
	if !found {
		return makeError(http.StatusNotFound, "no city found for CartoDB_ID %d", cityID)
	}

	distParam := r.URL.Query()["dist"]
	if len(distParam) == 1 {
		return getCitiesInBox(&f, distParam[0])
	}

	response = map[string]city{"city": featureToCity(f)}
	return
}

func getCitiesInBox(f *Feature, distParam string) (response interface{}, status int) {
	dist, err := strconv.ParseFloat(distParam, 64)
	if err != nil {
		return makeError(http.StatusBadRequest, "%q is not a valid distance", distParam)
	}

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

func featureToCity(f Feature) city {
	return city{
		CartoID:     f.Properties.CartoDBId,
		Name:        f.Properties.Name,
		Population:  f.Properties.Population,
		Coordinates: f.Geometry.Coordinates,
	}
}

type apiError struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func makeError(status int, format string, a ...interface{}) (interface{}, int) {
	return apiError{Error: fmt.Sprintf(format, a), Status: status}, status
}

func makeOk(a interface{}) (interface{}, int) {
	return a, http.StatusOK
}
