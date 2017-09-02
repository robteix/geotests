geotests
========

This repository is just me playing with geolocalization.


geotests is an API server that allows searching for specific cities by a
cartodb id as well as cities within a bounding box centred at a specific
city.

## Usage

    geotests [options]
    -exclude-origin
        Exclude point of origin of bounding box in results
    -f string
        A geojson file containing the data (shorthand) (default "data/canada_cities.geojson")
    -filename string
        A geojson file containing the data (default "canada_cities.geojson")
    -l string
        Where the server will listen to (default ":8000")
    -nz
        Ignore features with 0 population
    -pretty
        Indent JSON responses

The default invokation --

    geotests
    
will set off the following series of events:

1. The file `data/canada_cities.geojson` will be read and parsed
2. The data will be indexed in a couple of b-tree structures
3. A new HTTP server will be started on port 8000

Other examples:

    geotests -f mydata.json -nz -pretty
    
This will use `mydata.json` instead as the data source. It will not
index any feature (city) whose population is equal to 0 and the API
will pretty-print JSON responses.

    geotests -f mydata.json -nz -pretty -l 10.43.13.5:8080

Same as above but the HTTP server will only listen to connections to
port 8080 and IP address 10.43.13.5

## API

All API requests are answered with a JSON response. In case of success, the
response will be 200 (OK). In the event of an error, the status will depend
on the error and you will get a JSON response with a message. e.g.

    {
        "status" : 404
        "error" : "could not load city with id 123"
    }

Note that the server add a header (`Api-Response-Time`) to all responses
indicating the time it took to process the requests:

    HTTP/1.1 200 OK
    Api-Response-Time: 27.546µs
    Content-Type: application/json
    Date: Sat, 02 Sep 2017 12:23:03 GMT
    Content-Length: 185

### Fetching a single city by its cartoDB ID

You can retrieve information about a city by making a `GET` request providing a `cartodb_id`

    GET /id/<cartodbid>

If a city is found by the provided `cartodb_id`, the response will be a JSON object like --

    {
        "city": {
            "cartodb_id": 15712,
            "name": "Québec",
            "population": 0,
            "coordinates": [
                -71.214695,
                46.812407
            ]
        }
    }

If the city isn't found, the response will be a 404 (NOT FOUND). e.g.

    HTTP/1.1 404 Not Found
    Api-Response-Time: 26.438µs
    Content-Type: application/json
    Date: Sat, 02 Sep 2017 12:37:53 GMT
    Content-Length: 73

    {
        "status": 404,
        "error": "no city found for CartoDB_ID [111]"
    }

Note that if the server was started with the `-nz` option, the response
will be a 404 for cities with population 0 even if they are present in the
source data file.

### Fetching all cities nearby

You can find all cities within a bounding box defined by a distance in kilometres:

    GET /id/<origin_id>?dist=<dist in km>
    
The server will compute the bounding box defined as the minimum box that can hold a
circle of radius `dist`, centred at the coordinates of the point of origin, which is
the city identified by `origin_id` (a CartoDB_ID). All cities within the resulting
bounding box will be returned in the following format --

    {
        "cities": {
            "15712": {
                "cartodb_id": 15712,
                "name": "Québec",
                "population": 0,
                "coordinates": [
                    -71.214695,
                    46.812407
                ]
            },
            "15738": {
                "cartodb_id": 15738,
                "name": "Escalier Badelard",
                "population": 0,
                "coordinates": [
                    -71.216271,
                    46.816677
                ]
            }
        }
    }

Note that by default, the server will include the point of origin in the results, i.e.
if you do --

    GET /id/15712?dist=0.2
    
The results *will* include the city with ID 15712 (Québec City, Canada) along with any
other feature found within 200m. This behaviour can be changed by specifying the option
`-exclude-origin` when starting the server.

For example, by default a request to `/id/15712?dist=0.2` will result in the example above,
but the same request using `-exclude-origin` will give us --

    {
        "cities": {
            "15738": {
                "cartodb_id": 15738,
                "name": "Escalier Badelard",
                "population": 0,
                "coordinates": [
                    -71.216271,
                    46.816677
                ]
            }
        }
    }
