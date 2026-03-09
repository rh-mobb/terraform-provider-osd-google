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

variable "role_prefix" {
  type        = string
  default     = null
  nullable    = true
  description = "Prefix for custom IAM roles in GCP. When null, uses cluster_name. Hyphens and underscores are stripped (OCM allows only alphanumeric)."
}

variable "cluster_name" {
  type        = string
  default     = "my-wif-cluster"
  description = "Name of the cluster"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z). Used for cluster and WIF; WIF roles use x.y only."
}

variable "admin_password" {
  type        = string
  sensitive   = true
  default     = ""
  description = "Cluster admin password (optional; auto-generated if omitted)"
}
