variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM offline token or access token"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "gcp_project_number" {
  type        = string
  description = "GCP project number (numeric)"
}

variable "psc_subnet" {
  type        = string
  description = "Subnet name or CIDR for PSC service attachment"
}

variable "cluster_name" {
  type        = string
  default     = "my-psc-cluster"
  description = "Name of the cluster"
}
