package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type city struct {
	CartoID     int64       `json:"cartodb_id"`
	Name        string      `json:"name"`
	Population  int         `json:"population"`
	Coordinates Coordinates `json:"coordinates"`
}

func setupAPIRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/id/{cityId}", getIDHandler).Methods("GET")

	return r
}

func getIDHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now() // for benchmarking
	idS := mux.Vars(r)["cityId"]
	cityID, err := strconv.ParseInt(idS, 10, 64)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest, start)
		return
	}

	f, found := featureCollection.FindID(cityID)
	if !found {
		sendError(w, fmt.Sprintf("no city found for CartoDB_ID %d", cityID), http.StatusNotFound, start)
		return
	}

	distParam := r.URL.Query()["dist"]
	if len(distParam) == 1 {
		getCitiesInBox(w, &f, distParam[0], start)
		return
	}

	c := featureToCity(f)

	sendOk(w, c, start)
}

func getCitiesInBox(w http.ResponseWriter, f *Feature, distParam string, start time.Time) {
	dist, err := strconv.ParseFloat(distParam, 64)
	if err != nil {
		sendError(w, fmt.Sprintf("%q is not a valid distance", distParam), http.StatusBadRequest, start)
		return
	}

	feats, err := featureCollection.GetFeaturesNear(f.Properties.CartoDBId, dist)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError, start)
		return
	}

	cities := struct {
		Cities []city `json:"cities"`
	}{
		Cities: make([]city, len(feats)),
	}

	for k, v := range feats {
		cities.Cities[k] = featureToCity(v)
	}

	sendOk(w, cities, start)
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
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, message string, status int, start time.Time) {
	sendResponse(w, apiError{Error: message}, status, start)
}

func sendOk(w http.ResponseWriter, response interface{}, start time.Time) {
	sendResponse(w, response, http.StatusOK, start)
}

func sendResponse(w http.ResponseWriter, response interface{}, status int, start time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("API-Elapsed-Time", fmt.Sprint(time.Since(start)))
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
