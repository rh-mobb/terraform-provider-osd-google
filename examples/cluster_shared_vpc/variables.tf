variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM offline token or access token"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP service project ID (where cluster resources are created)"
}

variable "gcp_project_number" {
  type        = string
  description = "GCP project number (numeric)"
}

variable "vpc_host_project_id" {
  type        = string
  description = "GCP host project ID (shared VPC owner)"
}

variable "vpc_name" {
  type        = string
  description = "Name of the shared VPC network"
}

variable "compute_subnet" {
  type        = string
  description = "Subnet name for worker nodes"
}

variable "control_plane_subnet" {
  type        = string
  description = "Subnet name for control plane nodes"
}

variable "cluster_name" {
  type        = string
  default     = "my-shared-vpc-cluster"
  description = "Name of the cluster"
}
