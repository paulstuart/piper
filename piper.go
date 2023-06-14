package piper

import (
	"log"

	"golang.org/x/exp/constraints"
)

var Debug bool

type Float = constraints.Float

func max[T Float](a, b T) T {
	if a > b {
		return a
	}

	return b
}

func min[T Float](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func between[T Float](p, min, max T) bool {
	return p > min && p < max
}

// InExtent creates a bounding box of the outer ring
// and returns true if the point is in that box
func InExtent[T Float](p []T, ring [][]T) bool {
	w := ring[0][0]
	s := ring[0][1]
	e := ring[0][0]
	n := ring[0][1]

	for _, p := range ring {
		if w > p[0] {
			w = p[0]
		}

		if s > p[1] {
			s = p[1]
		}

		if e < p[0] {
			e = p[0]
		}

		if n < p[1] {
			n = p[1]
		}
	}
	lon, lat := p[0], p[1]

	return (((w <= lon) && (lon <= w)) ||
		((e <= lon) && (lon <= w)) ||
		((s <= lat) && (lat <= n)) ||
		((n <= lat) && (lat <= s)))
}

func InRing[T Float](p []T, ring [][]T) bool {
	first, last := ring[0], ring[len(ring)-1]
	if first[0] == last[0] && first[1] == last[1] {
		ring = ring[0 : len(ring)-1]
	}

	lon := p[0]
	lat := p[1]
	counter := 0

	for i, j := 0, len(ring)-1; i < len(ring); i, j = i+1, i {
		iLon := ring[i][0]
		iLat := ring[i][1]
		jLon := ring[j][0]
		jLat := ring[j][1]

		// if p's latitude is not between the edges latitudes it cannot intersect
		min := min(iLat, jLat)
		max := max(iLat, jLat)
		if !between(p[1], min, max) {
			continue
		}

		// if p's longitude is smaller than the longitude of the
		// rays intersection with the current edge then it intersects
		intersects := lon < ((jLon-iLon)*(lat-iLat))/(jLat-iLat)+iLon
		if intersects {
			counter++
			if Debug {
				log.Printf("polygon ray across [%d:%d]: %v -> %v", i, j, ring[i], ring[j])
			}
		}
	}

	if counter > 0 {
		if Debug {
			log.Printf("polygon ray crossed: %d segments", counter)
		}
	}
	return counter%2 != 0
}

func hasHoles[T Float](polygon [][][]T) bool {
	return len(polygon) > 1
}

// PipBox checks if the point falls in the bounding box of the polygon
// before actually checking the polygon
// This speeds up operations on complex polygons, insignifically slows
// down on simple polygons
func PipBox[T Float](p []T, polygon [][][]T) bool {
	if !InExtent(p, polygon[0]) {
		return false
	}
	return Pip(p, polygon)
}

// Pip checks if Point p is inside input polygon. Does account for holes.
func Pip[T Float](p []T, polygon [][][]T) bool {
	if Debug {
		log.Printf("PIP PT:%v in ring:%v", p, polygon[0][0])
	}
	if InRing(p, polygon[0]) {
		if Debug {
			log.Printf("InRing: %v", p)
		}
		// if there inner ring/holes we have to assume
		// that p can be in a hole, and therefor not in polygon
		if hasHoles(polygon) {
			holes := polygon[1:]

			for i := 0; i < len(holes); i++ {
				if InRing(p, holes[i]) {
					log.Printf("InHole [%d:%d]: %v", i, len(holes), p)
					return false
				}
			}
		}

		return true
	}

	return false
}
