variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region (e.g., us-central1)"
}

variable "cluster_name" {
  type        = string
  description = "Name of the OSD cluster (used for resource naming)"
}

variable "routing_mode" {
  type        = string
  description = "VPC routing mode: REGIONAL or GLOBAL"
  default     = "REGIONAL"
}

variable "master_cidr" {
  type        = string
  description = "CIDR for control plane subnet"
  default     = "10.0.0.0/19"
}

variable "worker_cidr" {
  type        = string
  description = "CIDR for worker/compute subnet"
  default     = "10.0.32.0/19"
}

variable "enable_psc" {
  type        = bool
  description = "Create PSC subnet and resources for private cluster"
  default     = false
}

variable "psc_cidr" {
  type        = string
  description = "CIDR for PSC subnet. Must be /29 or larger and within machine CIDR range."
  default     = "10.0.64.0/29"
}

variable "enable_private_cluster" {
  type        = bool
  description = "Create firewall rules for private cluster (master/worker internal, PSC if enabled)"
  default     = false
}

variable "enable_bastion_access" {
  type        = bool
  description = "Create firewall rule for bastion-to-cluster access (requires enable_private_cluster)"
  default     = false
}

variable "bastion_cidr" {
  type        = string
  description = "CIDR for bastion subnet (used when enable_bastion_access is true)"
  default     = "10.0.128.0/24"
}
