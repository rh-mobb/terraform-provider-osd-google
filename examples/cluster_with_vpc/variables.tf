variable "gcp_project_id" {
  type        = string
  description = "GCP project ID"
}

variable "cluster_name" {
  type        = string
  default     = "my-vpc-cluster"
  description = "Cluster name"
}

variable "gcp_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region"
}

variable "enable_psc" {
  type        = bool
  default     = false
  description = "Enable PSC for private cluster (requires OpenShift 4.17+)"
}
