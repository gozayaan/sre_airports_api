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
