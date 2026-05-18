variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (optional; use OSDGOOGLE_TOKEN env var instead)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "cluster_name" {
  type        = string
  default     = "my-cluster"
  description = "Cluster name. WIF config is looked up as {cluster_name}-wif."
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "cloud_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region for the cluster and KMS key ring"
}

variable "compute_nodes" {
  type        = number
  default     = 3
  description = "Number of default worker nodes"
}

variable "wait_for_create_complete" {
  type        = bool
  default     = true
  description = "Wait for cluster creation to complete before marking resource created"
}

variable "wait_timeout" {
  type        = number
  default     = 60
  description = "Timeout in minutes when waiting for cluster creation"
}
