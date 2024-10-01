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