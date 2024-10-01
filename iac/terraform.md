# Terraform Setup

> 💡 NOTE: First create a new project in GCP and _export_ the `GOOGLE_CLOUD_PROJECT` environment variable as the _project ID_ and `TF_VAR_BUCKET_NAME` variable as the _GCS bucket name_.

Inspect the terraform configuration in [`./gcs-bucket-provision/main.tf`](./gcs-bucket-provision/main.tf) and [`vars.tf`](./gcs-bucket-provision/vars.tf).

#### Sample main.tf

```terraform
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
```

#### Sample vars.tf

```terraform
variable "bucket_name" {
  description = "name of the bucket."
  type        = string
  default     = "bd-airport-data"
}

variable "storage_class" {
  description = "Bucket storage class."
  type        = string
  default     = "STANDARD"
}

variable "project_id" {
  description = "Bucket project id."
  type        = string
}

variable "location" {
  description = "Bucket location."
  type        = string
  default     = "US-CENTRAL1"
}
```

#### TF Configuration

- Initialize the GCP provider and setup the project with `terraform init`.

- Preview the execution plan with `terraform plan`.

- Double-check resource states while applying `terraform apply` and provision the **STANDARD** storage class bucket in the **US** region.
- Inspect currently applied resource state with `terraform show`.
