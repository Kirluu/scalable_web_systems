package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang/geo/s2"
	"google.golang.org/appengine/urlfetch"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"strconv"
)

func searchCountry(w http.ResponseWriter, ctx context.Context, country string, timeArg1 string, timeArg2 string) {
	client := urlfetch.Client(ctx)

	regions := [...]string{"europe", "north-america", "south-america", "asia", "central-america", "australia-ocenania", "africa", "antarctica"}

	for _, region := range regions {
		resp, err := client.Get(fmt.Sprintf("http://download.geofabrik.de/%s/%s.poly", region, country))
		if err == nil {
			bla := parseGeofabrikResponse(w, resp)

			bla2 := handlePolygon(w, bla)

			countImages(ctx, w, bla2, timeArg1, timeArg2)
			return
		}
	}

}

func parseGeofabrikResponse(w http.ResponseWriter, resp *http.Response) [][2]float64 {
	res := make([][2]float64, 0) // 80 because optimized for Denmark query ;)

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
			parsed, err := ParseFloat(w, elem)
			if err == nil {
				elemRes[elemResIndex] = parsed
				elemResIndex++
			}
			// Stop once we have inserted two elements - if there are more, ignore them
			if elemResIndex == 2 {
				break
			}
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
	res := make([][4]float64, 0)

	// Convert to s2.Point:
	points := make([]s2.Point, 0)
	for _, floatPair := range geofabrikResult {

		latLng := s2.LatLngFromDegrees(floatPair[0], floatPair[1])
		point := s2.PointFromLatLng(latLng)

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
		res = append(res, [4]float64{c.RectBound().Lo().Lat.Degrees(), c.RectBound().Hi().Lat.Degrees(), c.RectBound().Lo().Lng.Degrees(), c.RectBound().Hi().Lng.Degrees()})

		totalArea += c.RectBound().Area()
	}
	fmt.Fprintf(w, "Amount of points retrieved from Geofabrik: %d\n", len(geofabrikResult))
	fmt.Fprintf(w, "Total Area of rectangles representing the country: %v\n", totalArea)

	return res
}

func countImages(ctx context.Context, w http.ResponseWriter, rectangles [][4]float64, time1 string, time2 string) {

	ctx, _ = context.WithTimeout(ctx, 1*time.Minute)
	client, err := bigquery.NewClient(ctx, "johaa-178408")
	if err != nil {
		fmt.Fprintf(w, "error when creating BigQuery client from appengine context!")
		return
	}

	queryString := "SELECT COUNT(distinct base_url) as COUNT FROM `bigquery-public-data.cloud_storage_geo_index.sentinel_2_index` where"

	first := true

	for _, rect := range rectangles {
		if !first {
			queryString += " or ( "
		} else {
			queryString += " ( "
			first = false
		}
		lat1 := rect[0]
		lat2 := rect[1]
		long1 := rect[2]
		long2 := rect[3]

		queryString += fmt.Sprintf(" north_lat < %g and west_lon > %g and south_lat > %g and east_lon < %g )", lat2, long1, lat1, long2)
		queryString += fmt.Sprintf(" or ((north_lat > %g and south_lat < %g) and (west_lon < %g and east_lon > %g) )", lat2, lat2, long1, long1)
	}

	if time1 != "" && time2 != "" {
		queryString += " and (generation_time > " + time1 + " and generation_time < " + time2 + " )"
	}

	query := client.Query(queryString)

	//fmt.Fprintf(w, "\n%s\n\n", queryString)

	// Use standard SQL syntax for queries.
	// See: https://cloud.google.com/bigquery/sql-reference/
	query.QueryConfig.UseStandardSQL = true

	var queryIterator, readErr = query.Read(ctx)
	if readErr != nil {
		fmt.Fprintf(w, "Error reading bigquery")
		return
	}

	for {
		var baseUrl []bigquery.Value
		err := queryIterator.Next(&baseUrl)
		if err == iterator.Done {
			fmt.Fprintf(w, "Didn't find a count")
			return
		}
		if err != nil {
			fmt.Fprintf(w, "Error reading bigquery")
			return
		}
		sumstr := fmt.Sprintf("%s", baseUrl[0])
		if strings.Contains(sumstr, "int64") {
			count := strings.Split(sumstr, "=")[1]
			fcount := strings.Trim(count, ")")
			fmt.Fprintf(w, "Found %s images", fcount)
			return
		}
	}
}

type MyCount struct {
	theCount int64
}
