1. Provision a Cloud Storage Bucket Using Infrastructure as Code (IaC)
We'll use Terraform to provision a Google Cloud Storage (GCS) bucket.

```
Terraform Script: main.tf
provider "google" {
  project = "your-gcp-project-id"   # Replace with your GCP project ID
  region  = "us-central1"      	# Replace with your preferred region
  credentials = file("~/secrets/faisal-terraform-key.json")
}

resource "google_storage_bucket" "airport_images" {
  name      	= "airport-images-bucket" # Unique name for your bucket
  location  	= "US"
  storage_class   = "STANDARD"

  versioning {
	enabled = true
  }

  lifecycle {
	prevent_destroy = true
  }
}
```

Instructions:

Run the following commands to provision the bucket:

```
terraform init
terraform plan
terraform apply
```




2. Make an Endpoint /update_airport_image to Update an Airport’s Image
Next, we’ll implement an endpoint to allow uploading images for airports in our Go application.
Go Application Code:

```
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
```



3. Containerize the Go Application

Next, we’ll create a Docker image for our Go application.

```
# Use the official Golang image as the base image
FROM golang:1.20 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Start a new stage from scratch
FROM gcr.io/distroless/base

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["/main"]
```


Instructions:

	Save the Dockerfile in the root of your Go application directory.
	Build the Docker image:
```
docker build -t gcr.io/your-gcp-project-id/airport-app:latest .
```


4. Prepare a Deployment and Service Resource to Deploy in Kubernetes
Now, let’s create Kubernetes resources (Deployment and Service) to deploy our containerized application.
Deployment YAML: deployment.yaml

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: airport-app
spec:
  replicas: 3
  selector:
	matchLabels:
  	app: airport-app
  template:
	metadata:
  	labels:
    	app: airport-app
	spec:
  	containers:
  	- name: airport-app
    	image: gcr.io/your-gcp-project-id/airport-app:latest
    	ports:
    	- containerPort: 8080
    	env:
      	- name: GOOGLE_APPLICATION_CREDENTIALS
        	value: "/path/to/your/service-account-key.json" # Adjust accordingly
    	volumeMounts:
      	- name: gcs-auth
        	mountPath: /path/to/your/service-account-key.json
        	subPath: service-account-key.json
  	volumes:
    	- name: gcs-auth
      	secret:
        	secretName: gcs-auth-secret  # Ensure this secret is created with your GCS credentials




Service YAML: service.yaml
apiVersion: v1
kind: Service
metadata:
  name: airport-app
spec:
  type: LoadBalancer
  ports:
	- port: 80
  	targetPort: 8080
  selector:
	app: airport-app


Instructions:

	Ensure you have kubectl configured to point to your GKE cluster.
	Apply the deployment and service configurations:

kubectl apply -f deployment.yaml
kubectl apply -f service.yaml



5. Use API Gateway to Create Routing Rules to Send 20% of Traffic to the /airports_v2 Endpoint
Google Cloud API Gateway can manage routing rules for your application.
Creating API Gateway Configuration:
Create an OpenAPI Spec File: Create a file named openapi.yaml.
openapi: "3.0.0"
info:
  title: Airport API
  description: API for managing airport data
  version: 1.0.0

paths:
  /airports:
	get:
  	operationId: getAirports
  	responses:
    	'200':
      	description: A list of airports
      	content:
        	application/json:
          	schema:
            	type: array
            	items:
              	$ref: '#/components/schemas/Airport'

  /airports_v2:
	get:
  	operationId: getAirportsV2
  	responses:
    	'200':
      	description: A list of airports with extended info
      	content:
        	application/json:
          	schema:
            	type: array
            	items:
              	$ref: '#/components/schemas/AirportV2'

components:
  schemas:
	Airport:
  	type: object
  	properties:
    	name:
      	type: string
    	city:
      	type: string
    	iata:
      	type: string
    	image_url:
      	type: string
	AirportV2:
  	type: object
  	properties:
    	name:
      	type: string
    	city:
      	type: string
    	iata:
      	type: string
    	image_url:
      	type: string
    	runway_length:
      	type: integer
```

Deploy the API Gateway:

```
gcloud api-gateway api-configs create config1 \
  --api=my-api \
  --openapi-spec=openapi.yaml \
  --project=your-gcp-project-id \
  --api-gateway=my-api-gateway
```

Set up traffic splitting using the gcloud command line. Here’s how:
Create an API Config: First, make sure you have your API config defined in an OpenAPI spec file (e.g., openapi.yaml).
Deploy the API Config: Use the following command to deploy your API with traffic splitting:
```
gcloud api-gateway api-configs create config1 \
  --api=my-api \
  --openapi-spec=openapi.yaml \
  --project=your-gcp-project-id \
  --api-gateway=my-api-gateway \
  --traffic-split='{"/airports": 80, "/airports_v2": 20}'
  ```


Deploy the Traffic Split: If you already have an API config, you can update it:

```
gcloud api-gateway api-configs update config1 \
  --api=my-api \
  --openapi-spec=openapi.yaml \
  --project=your-gcp-project-id \
  --api-gateway=my-api-gateway \
  --traffic-split='{"/airports": 80, "/airports_v2": 20}'
```



1. Test the Endpoints
You can test the endpoints by sending requests and monitoring the responses. Below are examples using curl and Postman.
Using curl:
You can send a series of requests to your API Gateway endpoint to check the distribution of traffic.
Send Requests to api gateway(expecting 80% of traffic):

```
for i in {1..10}; do
  curl -X GET https://your-api-gateway-url/
done
```







