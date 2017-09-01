geotests
========

This repository is just me playing with geolocalization.


geotests is an API server that allows searching for specific cities by a
cartodb id as well as cities within a bounding box centred at a specific
city.

Usage:

    geotests [options]

    -f string
        A geojson file containing the data (shorthand) (default "canada_cities.geojson")
    -filename string
        A geojson file containing the data (default "canada_cities.geojson")
    -l string
        Where the server will listen to (default ":8000")
    -nz
        Ignore features with population 0

API
---

All API requests are answered with a JSON response. In case of success, the
response will be 200 (OK). In the event of an error, the status will depend
on the error and you will get a JSON response with a message. e.g.

    {
        "error": "could not load city with id 123"
    }

*Fetching a single city by its cartoDB ID*

    GET /id/<cartodbid>

