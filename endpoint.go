package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

type Airport struct {
	Name 	string `json:"name"`
	City 	string `json:"city"`
	IATA 	string `json:"iata"`
	ImageURL string `json:"image_url"`
}

var airports = []Airport{
	{"Hazrat Shahjalal International Airport", "Dhaka", "DAC", "https://storage.googleapis.com/bd-airport-data/dac.jpg"},
	{"Shah Amanat International Airport", "Chittagong", "CGP", "https://storage.googleapis.com/bd-airport-data/cgp.jpg"},
	{"Osmani International Airport", "Sylhet", "ZYL", "https://storage.googleapis.com/bd-airport-data/zyl.jpg"},
}

func UpdateAirportImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
    	http.Error(w, "Error parsing form data", http.StatusBadRequest)
    	return
	}

	airportName := r.FormValue("name")
	if airportName == "" {
    	http.Error(w, "Airport name is required", http.StatusBadRequest)
    	return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
    	http.Error(w, "Error retrieving the file", http.StatusBadRequest)
    	return
	}
	defer file.Close()

	var airport *Airport
	for i := range airports {
    	if airports[i].Name == airportName {
        	airport = &airports[i]
        	break
    	}
	}

	if airport == nil {
    	http.Error(w, "Airport not found", http.StatusNotFound)
    	return
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
    	http.Error(w, "Failed to create GCS client", http.StatusInternalServerError)
    	return
	}
	defer client.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
    	http.Error(w, "Error reading file", http.StatusInternalServerError)
    	return
	}

	bucketName := "airport-images-bucket" // Replace with your GCS bucket name
	objectName := fmt.Sprintf("airports/%s.jpg", airport.IATA)
	writer := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	writer.ContentType = header.Header.Get("Content-Type")

	if _, err := writer.Write(buf.Bytes()); err != nil {
    	http.Error(w, "Failed to upload image to GCS", http.StatusInternalServerError)
    	return
	}

	if err := writer.Close(); err != nil {
    	http.Error(w, "Failed to close the writer", http.StatusInternalServerError)
    	return
	}

	airport.ImageURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	response := map[string]string{
    	"message":   "Image updated successfully",
    	"image_url": airport.ImageURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/update_airport_image", UpdateAirportImage)
	http.ListenAndServe(":8080", nil)
}
