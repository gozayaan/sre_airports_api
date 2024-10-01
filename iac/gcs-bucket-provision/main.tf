# Create new GCS bucket in the US DC with Standard Storage
resource "google_storage_bucket" "static_bucket" {
 name          = var.bucket_name
 project       = var.project_id
 location      = var.location
 storage_class = var.storage_class
 uniform_bucket_level_access = true
}

# Upload a sample image as an object to GCS bucket
resource "google_storage_bucket_object" "static_object" {
 name         = "file.txt"
 source       = "/g/sample/file/path/file.jpeg"
 content_type = "text/plain"
 bucket       = google_storage_bucket
}