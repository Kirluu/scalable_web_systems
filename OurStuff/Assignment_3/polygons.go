package main

import (
	"fmt"
	"github.com/golang/geo/s2"
	"net/http"
	"context"
	"google.golang.org/appengine/urlfetch"
	"strings"
	"io/ioutil"
	"strconv"
)

func searchCountry(w http.ResponseWriter, ctx context.Context, country string, timeArg1 string, timeArg2 string) (int, error) {
	client := urlfetch.Client(ctx)

	regions := [...]string{"europe", "north-america", "south-america", "asia", "central-america", "australia-ocenania", "africa", "antarctica"}

	for _, region := range regions {
		resp, err := client.Get(fmt.Sprintf("http://download.geofabrik.de/%s/%s.poly", region, country))
		if (err != nil) {
			return dummy(ctx, handlePolygon(w, parseGeofabrikResponse(resp)), timeArg1, timeArg2) // <-- TODO: Call JARL method instead of dummy
		}
	}

	// Return bad result (-1) (Could've used error, but no need, since only the http-gets can fail, and these just mean the country couldn't be found)
	return -1, nil
}

func dummy(ctx context.Context, param [][4]float64, timeArg1 string, timeArg2 string) (int, error){
	return 1, nil
}

func parseGeofabrikResponse(resp *http.Response) [][2]float64 {
	res := make([][2]float64, 80) // 80 because optimized for Denmark query ;)

	body, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(body)

	lines := strings.Split(bodyString, "\n")

	for _, line := range lines {
		// Try to parse two float64 values in this line
		lineElems := strings.Split(line, "   ")
		var elemRes [2]float64
		elemResIndex := 0

		// Interpret the two first values in a line as a new lat/long combination
		for _, elem := range lineElems {
			parsed, err := strconv.ParseFloat(elem, 64)
			if err != nil {
				elemRes[elemResIndex] = parsed
				elemResIndex++
			}
			// Stop once we have inserted two elements - if there are more, ignore them
			if elemResIndex == 2 { break }
		}

		// Check that we successfully parsed two floats on the line: Only add as result if so
		if elemResIndex == 2 {
			res = append(res, elemRes)
		}
	}

	return res
}

// Return format for each array of size 4 => lat1, lat2, long1, long2
func handlePolygon(w http.ResponseWriter, geofabrikResult [][2]float64) [][4]float64 {
	res := make([][4]float64, 80)

	// Convert to s2.Point:
	points := make([]s2.Point, 80)
	for _, floatPair := range geofabrikResult {
		point := s2.PointFromLatLng(s2.LatLngFromDegrees(floatPair[0], floatPair[1]))
		points = append(points, point)
	}

	l1 := s2.LoopFromPoints(points)
	loops := []*s2.Loop{l1}
	poly := s2.PolygonFromLoops(loops)
	rc := &s2.RegionCoverer{MaxLevel: 30, MaxCells: 100}
	cover := rc.Covering(poly)
	var c s2.Cell
	var totalArea float64 = 0
	for i := 0; i < len(cover); i++ {
		c = s2.CellFromCellID(cover[i])
		// Store result-set with the 4 values representing the two ranges for lat and long
		res = append(res, [4]float64{c.RectBound().Lat.Lo, c.RectBound().Lat.Hi, c.RectBound().Lng.Lo, c.RectBound().Lng.Hi})

		totalArea += c.RectBound().Area()
	}
	fmt.Fprintf(w, "Amount of points retrieved from Geofabrik: %g", len(geofabrikResult))
	fmt.Fprintf(w, "Total Area of rectangles representing the country: %v", totalArea)

	return res
}