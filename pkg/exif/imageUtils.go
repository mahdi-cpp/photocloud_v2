package asset_create

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/draw"
	"math"
	"os"
)

var root = "var/cloud/"

type City struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	ProvinceId int    `json:"province_id"`
}

var cities []City
var names []string

func GetImageDimension(imagePath string) (int, int) {
	img, err := imaging.Open(imagePath) // Replace "image.jpg" with the path to your image file
	if err != nil {
		fmt.Println("Error opening image:", err)
		return 0, 0
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	fmt.Printf("Image width: %d\n", width)
	fmt.Printf("Image height: %d\n", height)
	return width, height
}

func GetNames() {

	filename := "var/cloud/data/name.txt"

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed after reading

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Read each line and append it to the slice
	for scanner.Scan() {
		names = append(names, scanner.Text())
	}

	// Check for any errors that occurred during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from file:", err)
		return
	}
}

func GetCities() {
	var folder = "/data/"
	var file = "cities.json"
	cities = []City{}

	f, err := os.Open(root + folder + file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cities); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
}

// CropImage crops an image to the specified rectangle.
func CropImage(src image.Image, rect image.Rectangle) image.Image {
	// Create a new blank image with the dimensions of the rectangle
	dst := image.NewRGBA(rect)
	// Draw the source image onto the destination image
	draw.Draw(dst, rect, src, rect.Min, draw.Over)
	return dst
}

// ProcessImage resizes and crops an image.
func ProcessImage(img image.Image, newWidth, newHeight int, cropRect image.Rectangle) (image.Image, image.Image) {
	// Resize the image
	resizedImage := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	// Crop the image
	croppedImage := imaging.Crop(resizedImage, cropRect)

	return resizedImage, croppedImage
}

func Dp(value float32) float32 {
	if value == 0 {
		return 0
	}
	return float32(math.Ceil(float64(2.625 * value)))
}
