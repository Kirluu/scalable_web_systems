package main

import (
	"fmt"
	"math"
	"github.com/golang/geo/s2"
	"net/http"
)

func searchCountry(client *http.Client, country string) {

	// TODO: Retrieve PSLG stuff


}

func handlePolygon(thingy [][2]float64) {
	// "Denmark Rectangle"
	p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(54.918, 8.552))
	p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(55.048, 8.471))
	p3 := s2.PointFromLatLng(s2.LatLngFromDegrees(55.481, 12.736))
	p4 := s2.PointFromLatLng(s2.LatLngFromDegrees(54.837, 9.392))
	p5 := s2.PointFromLatLng(s2.LatLngFromDegrees(54.918, 8.552))
	// synthetic example
	//p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(1, 1))
	//p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(2, 1))
	//p3 := s2.PointFromLatLng(s2.LatLngFromDegrees(2, 2))
	//p4 := s2.PointFromLatLng(s2.LatLngFromDegrees(1, 2))
	//p5 := s2.PointFromLatLng(s2.LatLngFromDegrees(1, 1))
	points := []s2.Point{p5, p4, p3, p2, p1}
	l1 := s2.LoopFromPoints(points)
	loops := []*s2.Loop{l1}
	poly := s2.PolygonFromLoops(loops)
	fmt.Printf("No. of edges %v\n", poly.NumEdges())
	// one big rectangle bounding box, just to test
	rect := poly.RectBound()
	fmt.Printf("Rect. Lat. Lo: %v \n", rect.Lat.Lo*180.0/math.Pi)
	fmt.Printf("Rect. Lat. Hi: %v \n", rect.Lat.Hi*180.0/math.Pi)
	fmt.Printf("Rect. Lng. Lo: %v \n", rect.Lng.Lo*180.0/math.Pi)
	fmt.Printf("Rect. Lng. Hi: %v \n", rect.Lng.Hi*180.0/math.Pi)
	fmt.Printf("\nOne Big Rect. Area %v\n\n", rect.Area())
	rc := &s2.RegionCoverer{MaxLevel: 30, MaxCells: 100}
	cover := rc.Covering(poly)
	var c s2.Cell
	var totalArea float64
	totalArea = 0
	for i := 0; i < len(cover); i++ {
		fmt.Printf("Cell %v : ", i)
		c = s2.CellFromCellID(cover[i])
		fmt.Printf("Low: %v - ", c.RectBound().Lo())
		fmt.Printf("High: %v \n", c.RectBound().Hi())
		totalArea = totalArea + c.RectBound().Area()
	}
	fmt.Printf("Total Area with multiple rectangles: %v", totalArea)
}