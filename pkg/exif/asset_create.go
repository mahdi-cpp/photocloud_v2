package asset_create

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/mahdi-cpp/PhotoKit/utils"
	"github.com/mahdi-cpp/photocloud_v2/pkg/asset_model"
	"image"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var uploadPath = "/var/cloud/upload/upload/"

const (
	AppDir        = "/media/mahdi/Cloud/apps/Photos/"
	AssetsDir     = "/assets/"
	MetadataDir   = "/metadata/"
	ThumbnailsDir = "/thumbnails/"
	IndexFile     = "index.dat"
)

var username = "mahdi_abdolmaleki"
var thumbnails = []int{270}

func CreateAssetOfUploadDirectory() {

	files, err := os.ReadDir(uploadPath)
	if err != nil {
		fmt.Println(err)
	}

	var idCounter = 0

	for _, file := range files {

		var fileNamed = file.Name()

		//fmt.Println("-----------:" + rootPath + username_path)

		//found, textFile, line, err := storage.SearchTextInFiles(rootPath+username_path + "assets/", ".json", fileNamed, true)
		//if err != nil {
		//	fmt.Printf("Error: %v\n", err)
		//	//continue
		//}
		//if found {
		//	fmt.Printf("Found text in %s (line %d)\n", textFile, line)
		//	continue
		//}

		if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".JPG") || strings.HasSuffix(file.Name(), ".jpeg") || strings.HasSuffix(file.Name(), ".JPEG") {

			//var assetUrl = uuid.New().String()
			//var assetFormat = ".jpg"

			err = CopyFile(uploadPath+file.Name(), AppDir+username+AssetsDir+strconv.Itoa(idCounter)+".jpg")
			if err != nil {
				panic(err)
			}

			//textFile, err := os.OpenFile(rootPath+username_path+assetUrl+".txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			//if err != nil {
			//	panic(err)
			//}
			//defer textFile.Close()
			//
			//if _, err := textFile.WriteString(file.Name() + "\n"); err != nil {
			//	panic(err)
			//}

			var portrait = false
			var Orientation = 0

			var assetPath = AppDir + username + AssetsDir + strconv.Itoa(idCounter) + ".jpg"

			//var cameraMake = ""
			//var cameraModel = ""

			var metadata ImageMetadata

			if PhotoHasExifData(assetPath) {

				has, orientation := ReadExifData(assetPath)

				metadata, err = GetImageMetadata(assetPath)

				if has {
					fmt.Println("Orientation: ", orientation)
					if strings.Compare(orientation, "6") == 0 {
						portrait = true
					}

					i, err := strconv.Atoi(orientation)
					if err != nil {
						fmt.Println("Orientation: ", err)
					} else {
						Orientation = i
					}
				}

				//cMake, cModel, err := utils.GetCameraModel(assetPath)
				//if err != nil {
				//	log.Printf("Warning: error getting camera info: %v", err)
				//	cameraMake = ""
				//	cameraModel = ""
				//} else {
				//	cameraMake = cMake
				//	cameraModel = cModel
				//
				//	// Convert to NULL if empty after sanitization
				//	if cameraMake == "" {
				//		cameraMake = "NULL" // For raw SQL, or use sql.NullString
				//	}
				//	if cameraModel == "" {
				//		cameraModel = "NULL"
				//	}
				//}

			} else {
				fmt.Println("not exif data")
			}

			w, h := getImageDimension(assetPath)
			var width = 0
			var height = 0
			if Orientation == 6 {
				width = h
				height = w
			} else {
				width = w
				height = h
			}

			var isAssetScreenshot = false
			if isScreenshot(fileNamed) {
				isAssetScreenshot = true
			}

			asset := asset_model.PHAsset{
				ID:          idCounter,
				UserID:      2,
				Url:         strconv.Itoa(idCounter),
				Filename:    file.Name(),
				MediaType:   "image",
				Format:      "jpg",
				Orientation: Orientation,

				CameraMake:  metadata.CameraMake,
				CameraModel: metadata.CameraModel,

				PixelWidth:  width,
				PixelHeight: height,

				IsScreenshot: isAssetScreenshot,

				CapturedDate: metadata.CaptureTime,
				CreationDate: time.Now(),
			}

			// Save the asset Metadata
			err = SaveAssetMetadata(asset)
			if err != nil {
				fmt.Println("Error saving asset:", err)
				return
			}

			idCounter++
			for _, tinySize := range thumbnails {
				CreateTinyAsset(file.Name(), tinySize, portrait, asset.ID)
			}

			//Save the chat to the database

			//if err := db.Debug().Create(&asset).Error; err != nil {
			//	log.Printf("Failed to create PHAsset: %v", err)
			//} else {
			//	fmt.Printf("Created PHAsset: %+v\n", asset)
			//
			//	for _, tinySize := range thumbnails {
			//		CreateTinyAsset(file.Name(), assetUrl, tinySize, portrait)
			//	}
			//}
		}
	}
}

func isScreenshot(filename string) bool {
	matched, _ := regexp.MatchString(`(?i)screenshot`, filename)
	return matched
}

func CreateOnlyDatabase(userId int) {

	var userIdPath = strconv.FormatInt(int64(userId), 10) + "/"

	files, err := os.ReadDir(AppDir + username)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {

		var named = strings.Replace(file.Name(), ".jpg", "", 1)

		// Check if asset exists by name
		//exists, err := checkUrlExists(named)
		//if err != nil {
		//	fmt.Printf("Error checking product: %v\n", err)
		//	continue
		//}

		//if exists {
		//	fmt.Printf("Asset '%s' exists\n", named)
		//	continue
		//}

		if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".JPG") || strings.HasSuffix(file.Name(), ".jpeg") || strings.HasSuffix(file.Name(), ".JPEG") {

			var Orientation = 0
			var a = AppDir + username + AssetsDir + file.Name()

			var cameraMake = ""
			var cameraModel = ""

			if PhotoHasExifData(a) {
				has, orientation := ReadExifData(a)

				if has {
					fmt.Println("Orientation: ", orientation)
					if strings.Compare(orientation, "6") == 0 {
						//portrait = true
					}

					i, err := strconv.Atoi(orientation)
					if err != nil {
						fmt.Println("Orientation: ", err)
					} else {
						Orientation = i
					}
				}

				cMake, cModel, err := utils.GetCameraModel(a)
				if err != nil {
					log.Printf("Warning: error getting camera info: %v", err)
					cameraMake = ""
					cameraModel = ""
				} else {
					cameraMake = cMake
					cameraModel = cModel

					// Convert to NULL if empty after sanitization
					if cameraMake == "" {
						cameraMake = "NULL" // For raw SQL, or use sql.NullString
					}
					if cameraModel == "" {
						cameraModel = "NULL"
					}
				}

			} else {
				fmt.Println("not exif data")
			}

			w, h := getImageDimension(AppDir + username + AssetsDir + userIdPath + file.Name())
			var width = 0
			var height = 0
			if Orientation == 6 {
				width = h
				height = w
			} else {
				width = w
				height = h
			}

			newPHAsset := asset_model.PHAsset{
				UserID:      userId,
				Url:         named,
				MediaType:   "image",
				Format:      "jpg",
				Orientation: Orientation,

				CameraMake:  cameraMake,
				CameraModel: cameraModel,

				PixelWidth:  width,
				PixelHeight: height,

				CreationDate: time.Now(),
			}

			fmt.Println(newPHAsset.UserID)

			//// Save the chat to the database
			//if err := db.Debug().Create(&newPHAsset).Error; err != nil {
			//	log.Printf("Failed to create PHAsset: %v", err)
			//}
		}
	}
}

func CreateTinyAsset(sourceName string, createSize int, portrait bool, id int) {

	file := uploadPath + sourceName
	fmt.Println("CreateTinyAsset: ", sourceName, createSize)

	srcImage, err := imaging.Open(file)
	if err != nil {
		panic(err)
	}

	var dstImage *image.NRGBA

	if portrait {
		// Resize the cropped image to width = 200px preserving the aspect ratio.
		dstImage = imaging.Resize(srcImage, 0, createSize, imaging.Lanczos)
		dstImage = imaging.Rotate270(dstImage)
	} else {
		// Resize the cropped image to width = 200px preserving the aspect ratio.
		dstImage = imaging.Resize(srcImage, createSize, 0, imaging.Lanczos)
	}

	//var name2 = AppDir + username + ThumbnailsDir + assetNewName + "_" + strconv.Itoa(createSize) + ".jpg"

	name2 := GetTinyPath(id)

	err = imaging.Save(dstImage, name2)
	if err != nil {
		panic(err)
	}
}

func GetTinyPath(id int) string {
	return filepath.Join(AppDir, username, ThumbnailsDir, fmt.Sprintf("%d_270.jpg", id))
}

func getImageDimension(imagePath string) (int, int) {
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
