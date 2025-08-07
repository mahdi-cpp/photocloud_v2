package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var polygonsDir = "/home/mahdi/photos/map/hamedan/"

// Polygon represents the structure of your JSON data
type Polygon struct {
	Name string        `json:"name"`
	Data [][][]float64 `json:"data"`
}

func polygon() {
	// Directory containing JSON files
	dir := polygonsDir

	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	// Process each JSON file
	for _, file := range files {

		// Skip non-JSON files
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Read file content
		filePath := filepath.Join(dir, file.Name())

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file.Name(), err)
			continue
		}

		// Unmarshal JSON into Polygon struct
		var polygon Polygon
		err = json.Unmarshal(content, &polygon)
		if err != nil {
			fmt.Printf("Error unmarshaling JSON from %s: %v\n", file.Name(), err)
			continue
		}

		// Print the polygon data
		fmt.Printf("Processed file: %s\n", file.Name())
		fmt.Printf("Polygon Name: %s\n", polygon.Name)
		fmt.Printf("Number of polygons: %d\n", len(polygon.Data))
		for i, poly := range polygon.Data {
			fmt.Printf("  Polygon %d has %d points\n", i+1, len(poly))
		}
		fmt.Println("---")
	}
}

// readPolygonsFromDirectory reads all JSON files from the given directory path,
// unmarshals each into a Polygon struct, and returns a slice of all valid polygons.
// It handles errors for individual files gracefully by logging and continuing.
func readPolygonsFromDirectory(dirPath string) ([]Polygon, error) {
	var polygons []Polygon

	// Read all entries (files and subdirectories) in the specified directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("could not read directory '%s': %w", dirPath, err)
	}

	for _, entry := range entries {
		// Skip directories and files that don't have a .json extension
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name()) // Construct the full path to the file

		// Read the content of the JSON file
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Could not read file '%s': %v\n", filePath, err)
			continue // Continue to the next file if reading fails
		}

		var poly Polygon
		// Unmarshal the JSON content into the Polygon struct
		err = json.Unmarshal(fileContent, &poly)
		if err != nil {
			fmt.Printf("Warning: Could not unmarshal JSON from '%s': %v\n", filePath, err)
			continue // Continue to the next file if unmarshalling fails
		}

		// If successful, add the polygon to our slice
		polygons = append(polygons, poly)
	}

	return polygons, nil
}

// Coordinate represents a geographic point with latitude and longitude
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

// IsCoordinateInPolygon checks if a point is inside a polygon using ray-casting algorithm
func IsCoordinateInPolygon(point Coordinate, polygon []Coordinate) bool {
	if len(polygon) < 3 {
		return false // Not a valid polygon
	}

	inside := false
	n := len(polygon)
	j := n - 1

	for i := 0; i < n; i++ {
		if rayIntersectsSegment(point, polygon[j], polygon[i]) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// rayIntersectsSegment checks if a ray from the point intersects with a polygon segment
func rayIntersectsSegment(p Coordinate, a Coordinate, b Coordinate) bool {
	// Check if point is between the segment's y-coordinates
	if (a.Latitude > p.Latitude) == (b.Latitude > p.Latitude) {
		return false
	}

	// Calculate x-intersection of the ray with the segment
	x := (b.Longitude-a.Longitude)*(p.Latitude-a.Latitude)/(b.Latitude-a.Latitude) + a.Longitude

	// Check if intersection is to the right of the point
	return x > p.Longitude
}

// IsCoordinateInBoundingBox checks if a point is within a rectangular bounding box
func IsCoordinateInBoundingBox(point Coordinate, minBound Coordinate, maxBound Coordinate) bool {
	return point.Latitude >= minBound.Latitude &&
		point.Latitude <= maxBound.Latitude &&
		point.Longitude >= minBound.Longitude &&
		point.Longitude <= maxBound.Longitude
}
