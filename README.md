# Airport API

<!-- My thought process and decisions goes here -->

---
_For tasks, checkout [tasks.md](tasks.md)_


// This is the updated code for deployment I have added the UpdateAirportImage handler to update the airport images section and custom bucket name. And changed port 9090. bcoz Jenkins running on 8080 port.






package main

import (
        "encoding/json"
        "log"
        "net/http"
        "strings"
        "io"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/s3"
)

type Airport struct {
        Name     string `json:"name"`
        City     string `json:"city"`
        IATA     string `json:"iata"`
        ImageURL string `json:"image_url"`
}

type AirportV2 struct {
        Airport
        RunwayLength int `json:"runway_length"`
}

var airports = []Airport{
        {"Hazrat Shahjalal International Airport", "Dhaka", "DAC", "https://storage.googleapis.com/bd-airport-data/dac.jpg"},
        {"Shah Amanat International Airport", "Chittagong", "CGP", "https://storage.googleapis.com/bd-airport-data/cgp.jpg"},
        {"Osmani International Airport", "Sylhet", "ZYL", "https://storage.googleapis.com/bd-airport-data/zyl.jpg"},
}

var airportsV2 = []AirportV2{
        {Airport{"Hazrat Shahjalal International Airport", "Dhaka", "DAC", "https://storage.googleapis.com/bd-airport-data/dac.jpg"}, 3200},
        {Airport{"Shah Amanat International Airport", "Chittagong", "CGP", "https://storage.googleapis.com/bd-airport-data/cgp.jpg"}, 2900},
        {Airport{"Osmani International Airport", "Sylhet", "ZYL", "https://storage.googleapis.com/bd-airport-data/zyl.jpg"}, 2500},
}

// HomePage handler
func HomePage(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Status: OK"))
}

// Airports handler
func Airports(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(airports)
}

// AirportsV2 handler
func AirportsV2(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(airportsV2)
}

// UpdateAirportImage handler for updating airport images
func UpdateAirportImage(w http.ResponseWriter, r *http.Request) {
        err := r.ParseMultipartForm(10 << 20) // Limit to 10 MB
        if err != nil {
                http.Error(w, "Unable to parse form", http.StatusBadRequest)
                return
        }

        airportName := r.FormValue("name")
        file, _, err := r.FormFile("image")
        if err != nil {
                http.Error(w, "Unable to get file", http.StatusBadRequest)
                return
        }
        defer file.Close()

        // Initialize AWS session
        sess, err := session.NewSession(&aws.Config{
                Region: aws.String("ap-southeast-1"), // Updated region
        })
        if err != nil {
                http.Error(w, "Unable to create AWS session", http.StatusInternalServerError)
                return
        }

        // Create S3 service client
        svc := s3.New(sess)

        // Define the path for the S3 object
        imgPath := strings.ReplaceAll(airportName, " ", "_") + ".jpg"

        // Upload the image to S3
        _, err = svc.PutObject(&s3.PutObjectInput{
                Bucket:      aws.String("amar-chobi"), // Updated bucket name
                Key:         aws.String(imgPath),
                Body:        file,
                ContentType: aws.String("image/jpeg"),
                ACL:        aws.String("public-read"), // Change as needed
        })
        if err != nil {
                http.Error(w, "Failed to upload to S3", http.StatusInternalServerError)
                return
        }

        // Construct the new image URL
        newImageURL := "https://amar-chobi.s3.ap-southeast-1.amazonaws.com/" + imgPath
        for i, airport := range airports {
                if airport.Name == airportName {
                        airports[i].ImageURL = newImageURL
                        break
                }
        }

        // Respond with success
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Image updated successfully", "image_url": newImageURL})
}

func main() {
        // Setup routes
        http.HandleFunc("/", HomePage)
        http.HandleFunc("/airports", Airports)
        http.HandleFunc("/airports_v2", AirportsV2) // Second endpoint
        http.HandleFunc("/update_airport_image", UpdateAirportImage)

        // Start the server
        log.Fatal(http.ListenAndServe(":9090", nil))
