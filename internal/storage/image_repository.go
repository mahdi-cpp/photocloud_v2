package storage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// https://chat.deepseek.com/a/chat/s/d3a14675-7d92-4bf8-a585-8e8db1fd0b62

var folders = []string{
	"/media/mahdi/Cloud/apps/Photos/ali_abdolmaleki/assets",
	"/media/mahdi/Cloud/apps/Photos/ali_abdolmaleki/thumbnails",

	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/assets",
	"/media/mahdi/Cloud/apps/Photos/mahdi_abdolmaleki/thumbnails",
}

//var iconFolder = "/var/cloud/icons/"

type ImageRepositoryV1 struct {
	sync.RWMutex
	memory map[string][]byte
}

//type IconRepository struct {
//	sync.RWMutex
//	memory map[string][]byte
//}

//var tinyRepository = ImageRepositoryV1{memory: make(map[string][]byte)}
//var iconRepository = ImageRepositoryV1{memory: make(map[string][]byte)}

//--------------------------------------------

func NewImageRepository() *ImageRepositoryV1 {
	r := &ImageRepositoryV1{
		memory: make(map[string][]byte),
	}

	r.loadIcons()
	return r
}

func (r *ImageRepositoryV1) GetImage(filename string) ([]byte, bool) {
	r.RLock()
	imgData, exists := r.memory[filename]
	r.RUnlock()
	return imgData, exists
}

//func (r *ImageRepositoryV1) GetIconCash(filename string) ([]byte, bool) {
//	iconRepository.RLock()
//	imgData, exists := iconRepository.cache[filename]
//	iconRepository.RUnlock()
//	return imgData, exists
//}

func (r *ImageRepositoryV1) SearchFile(filename string) (string, error) {
	for _, folder := range folders {
		// Construct the full path to the file
		fullPath := filepath.Join(folder, filename)

		// Check if the file exists
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil // File found
		} else if os.IsNotExist(err) {
			// File does not exist in this directory
			continue
		} else {
			// Other error (e.g., permission issues)
			return "", err
		}
	}
	return "", fmt.Errorf("file %s not found in any of the specified folders", filename)
}

func (r *ImageRepositoryV1) AddTinyImage(filepath string, filename string) {

	originalImage, err := r.loadImage(filepath)
	if err != nil {
		fmt.Println("addToCash Error loading image:", err)
		return
	}

	imgBytes, err := r.convertImageToBytes(originalImage, "jpg") // Change to "png" for PNG format
	if err != nil {
		fmt.Println("Error convertImageToBytes: ", err)
		return
	}

	r.Lock()
	r.memory[filename] = imgBytes
	r.Unlock()
}

// loadImage loads an image from a file.
func (r *ImageRepositoryV1) loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (r *ImageRepositoryV1) loadIcons() {

	// Specify the directory you want to read
	dir := "/var/cloud/icons" // Change this to your target directory

	// Read the directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read directory: %v", err)
	}

	// Iterate over the entries
	for _, entry := range entries {
		if !entry.IsDir() { // Check if it is not a directory

			if strings.Contains(entry.Name(), ".png") {
				r.addIconCash(entry.Name())
				//fmt.Printf("Reading file: %s\n", entry.Filename())
			}
		}
	}

	fmt.Println(len(r.memory))
}

// convertImageToBytes converts an image.Image to a byte slice.
func (r *ImageRepositoryV1) convertImageToBytes(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	// Encode the image based on the specified format
	switch format {
	case "jpg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *ImageRepositoryV1) addIconCash(iconName string) {
	icon, err := r.loadImage("/var/cloud/icons/" + iconName) // Change path accordingly
	if err != nil {
		fmt.Println("Error loading image:", err)
		return
	}

	imgBytes, err := r.convertImageToBytes(icon, "png") // Change to "png" for PNG format
	if err != nil {
		fmt.Println("Error convertImageToBytes: ", err)
		return
	}

	r.Lock()
	r.memory[iconName] = imgBytes
	r.Unlock()
}

//IDGenerator--------------------------------

var IdGen = NewIDGenerator()

// IDGenerator is a struct that holds the current ID and a mutex for thread safety
type IDGenerator struct {
	currentID int
	mu        sync.Mutex
}

// NewIDGenerator creates a new IDGenerator instance
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		currentID: 0,
	}
}

// NextID generates the next unique ID
func (g *IDGenerator) NextID() int {
	g.mu.Lock()         // Lock to ensure thread safety
	defer g.mu.Unlock() // Unlock after generating the ID
	g.currentID++       // Increment the current ID
	return g.currentID  // Return the new ID
}

// utils-------------------------------------

func GetFileSize(filepath string) (int64, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
