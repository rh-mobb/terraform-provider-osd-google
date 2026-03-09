variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "cluster_name" {
  type        = string
  default     = "my-cluster-with-admin"
  description = "Name of the cluster"
}

variable "admin_password" {
  type        = string
  sensitive   = true
  description = "Cluster admin password (optional; auto-generated if omitted)"
  default     = ""
}
