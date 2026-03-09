variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM token (optional; set token = var.ocm_token in provider to use)"
  default     = ""
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "cluster_name" {
  type        = string
  default     = "my-osd-cluster"
  description = "Name of the cluster"
}
