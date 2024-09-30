package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Airport struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	IATA    string `json:"iata"`
	ImageURL string `json:"image_url"`
}

type AirportV2 struct {
	Airport
	RunwayLength int `json:"runway_length"`
}

// Mock data for airports in Bangladesh
var airports = []Airport{
	{"Hazrat Shahjalal International Airport", "Dhaka", "DAC", "https://storage.googleapis.com/bd-airport-data/dac.jpg"},
	{"Shah Amanat International Airport", "Chittagong", "CGP", "https://storage.googleapis.com/bd-airport-data/cgp.jpg"},
	{"Osmani International Airport", "Sylhet", "ZYL", "https://storage.googleapis.com/bd-airport-data/zyl.jpg"},
}

// Mock data for airports in Bangladesh (with runway length for V2)
var airportsV2 = []AirportV2{
	{Airport{"Hazrat Shahjalal International Airport", "Dhaka", "DAC", "https://storage.googleapis.com/bd-airport-data/dac.jpg"}, 3200},
	{Airport{"Shah Amanat International Airport", "Chittagong", "CGP", "https://storage.googleapis.com/bd-airport-data/cgp.jpg"}, 2900},
	{Airport{"Osmani International Airport", "Sylhet", "ZYL", "https://storage.googleapis.com/bd-airport-data/zyl.jpg"}, 2500},
}
 
const (
	maxUploadFileSize = 10 * 1024 * 1024 // 10 MB
)


// HomePage handler
func HomePage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Status: OK"))
}

// Airports handler for the first endpoint
func Airports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airports)
}

// AirportsV2 handler for the second version endpoint
func AirportsV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airportsV2)
}

// UpdateAirportImage handler for updating airport images
func UpdateAirportImage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the airport name and image data
	switch r.Method {

	case http.MethodPost:
		// parse the form having two fields: airport_name and airport_img (contains image file)		
		r.ParseMultipartForm(maxUploadFileSize) 
		
		air_name := r.FormValue("airport_name")
		if air_name == "" {
			http.Error(w, "Empty Airport name", http.StatusBadRequest)
			return
		}

		fmt.Printf("\n🌍 Uploaded Airport Name: %+v\n", air_name)

		// parse the image data
		imageFile, header, err := r.FormFile("airport_img")
        if err != nil {
            log.Println("Image file retrieval Error", err)
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

		fileExt := filepath.Ext(header.Filename)
        defer imageFile.Close()
		
		// Find the airport by name
		var flag bool = false
		var indexV2 int = 0 		
		var matched_airport *Airport

		for x := range airports {
			if air_name == airports[x].Name {
				flag = true
				indexV2 = x
				matched_airport = &airports[x]
				break
			}
		}
	
		if !flag && matched_airport == nil {
			fmt.Printf("No airport matched for %v\n", air_name)
			http.Error(w, "No airport matched", http.StatusNotFound)
			return
		}

		
		// Initialize GCS client with mock GCS server
		gcsURL := os.Getenv("GCS_LOCALHOST_URL")
		gcsclient, err := storage.NewClient(context.TODO(), option.WithEndpoint(gcsURL))
		if err != nil {
			log.Fatalf("failed to create GCS client: %v", err)
			http.Error(w, "GCS client failure", http.StatusInternalServerError)
		}
		defer gcsclient.Close()
		
		var objectList []string
		
		bucketName := os.Getenv("GCS_BUCKET_NAME")
		
		// iterate over existing bucket contents
		x := gcsclient.Bucket(bucketName).Objects(context.Background(), &storage.Query{})
		for {
			oattribs, err := x.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("failed to list: %v", err)
			}
			// fmt.Printf("Adding Object: %+v\n", oattribs.Name)
			objectList = append(objectList, oattribs.Name)
		}
		fmt.Printf("🧺 Bucket objects: %+v\n", objectList)

		
		// Upload image to GCS and update the airport's image URL

		// set resource release before timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
		defer cancel()

		// filename format: <Airport-IATA>.<EXT>
		fileNameWithExt := fmt.Sprintf("%s%s\n",strings.ToLower(matched_airport.IATA), fileExt)

		// Upload an object with storage.Writer.
		writeClient := gcsclient.Bucket(bucketName).Object(fileNameWithExt).NewWriter(ctx)
		if _, err := io.Copy(writeClient, imageFile); err != nil {
			log.Printf("failed to upload file ❌, %v", err)
		}
		
		fmt.Printf("File uploaded to bucket! ✅\n")

		if err := writeClient.Close(); err != nil {
			log.Fatal(err)
			error := fmt.Errorf("Writer.Close: %v", err)
			fmt.Println(error.Error())
		}	

		// Update existing URLs
		// URL spec : https://storage.googleapis.com/<bucket-name>/<object-name>

		bucketDomain := os.Getenv("GCS_BUCKET_DOMAIN")
		uploadedImgURL := fmt.Sprintf("https://%s/%s/%s", bucketDomain,bucketName, fileNameWithExt)
		matched_airport.ImageURL = uploadedImgURL
		airportsV2[indexV2].Airport.ImageURL = uploadedImgURL


		// Respond with success/failure
		fmt.Printf("🔸 V1 API data - Airport City:%v IATA:%v URL:%v", matched_airport.City, matched_airport.IATA, matched_airport.ImageURL)
		fmt.Printf("🔸 V2 API data - Airport City:%v IATA:%v URL:%v\n", airportsV2[indexV2].Airport.City, airportsV2[indexV2].Airport.IATA, airportsV2[indexV2].Airport.ImageURL,)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Image uploaded",
			"filename": fileNameWithExt,
			"image_url": uploadedImgURL,
		})

	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodDelete:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// Setup routes
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/airports", Airports)
	http.HandleFunc("/airports_v2", AirportsV2)
	http.HandleFunc("/update_airport_image", UpdateAirportImage)

	// Start the server
	http.ListenAndServe(":8080", nil)
}