# Create new GCS bucket in the US DC with Standard Storage

resource "google_storage_bucket" "static" {
 name          = "bd-airport-data"
 location      = "US"
 storage_class = "STANDARD"
 
 uniform_bucket_level_access = true
}

# Upload a sample image as an object to GCS bucket

resource "google_storage_bucket_object" "default" {
 name         = "file.txt"
 source       = "/g/sample/file/path/file.jpeg"
 content_type = "text/plain"
 bucket       = google_storage_bucket
}