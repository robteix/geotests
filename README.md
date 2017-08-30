geotests
========

This repository is just me playing with geolocalization.

boundarybox
-----------

Implements the algorithm to return a boundary box defined as the minimum box
that holds a circle projected on a sphere (the Earth) centred at given
coordinates and radius in kilometres.

Usage:

  boundingbox [options]

  -dist float
        distance (radius) in Km (default 2)
  -lat float
        latitude (default 46.716993)
  -lon float
        longitude (default -71.269204)

(The defaults are the city of Charny, near Quebec City.)

Examples:

    $ ./boundingbox -lon 180 -lat 90 -dist 1000
    81.006798,-180.000000,90.000000,180.000000
    $ ./boundingbox -lon 78 -lat 180 -dist 1000
    171.006798,-180.000000,90.000000,180.000000
    $./boundingbox -lon 78 -lat 180 -dist 5000
    135.033990,-180.000000,90.000000,180.000000

