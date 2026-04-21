# Required
variable "name" {
  type        = string
  description = "Cluster name (must match WIF config display_name prefix; WIF lookup uses {name}-wif)"
}

variable "cloud_region" {
  type        = string
  description = "GCP region (e.g., us-central1)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

# WIF lookup
variable "wif_config_display_name" {
  type        = string
  default     = null
  nullable    = true
  description = "WIF config display name for data source lookup. Defaults to {name}-wif."
}

# Cluster options
variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (e.g., 4.21.3)"
}

variable "compute_nodes" {
  type        = number
  default     = 3
  description = "Number of default worker nodes (without autoscaling)"
}

variable "compute_machine_type" {
  type        = string
  default     = null
  description = "GCP machine type for default worker pool (e.g., n2-standard-4). Omit for provider default."
}

variable "ccs_enabled" {
  type        = bool
  default     = true
  description = "Customer Cloud Subscription (required for WIF)"
}

variable "multi_az" {
  type        = bool
  default     = null
  description = "Deploy across multiple availability zones"
}

variable "availability_zones" {
  type        = list(string)
  default     = null
  description = "GCP availability zones. 1 for single-AZ, 3 for multi-AZ. Omit for provider default."
}

variable "domain_prefix" {
  type        = string
  default     = null
  description = "DNS domain prefix for the cluster"
}

variable "billing_model" {
  type        = string
  default     = null
  description = "Billing model: standard or marketplace-gcp"
}

variable "properties" {
  type        = map(string)
  default     = null
  description = "Cluster properties"
}

# Private cluster
variable "private" {
  type        = bool
  default     = false
  description = "Restrict cluster API endpoint to private (internal) listening only. Requires BYO VPC (gcp_network) and PSC (private_service_connect). Cannot be changed after creation."
}

# Network
variable "gcp_network" {
  type = object({
    vpc_name             = string
    compute_subnet       = string
    control_plane_subnet = string
    vpc_project_id       = optional(string)
  })
  default     = null
  description = "BYO VPC: vpc_name, compute_subnet, control_plane_subnet, optional vpc_project_id"
}

variable "private_service_connect" {
  type = object({
    service_attachment_subnet = string
  })
  default     = null
  description = "Private Service Connect config (service_attachment_subnet required for PSC)"
}

variable "security" {
  type = object({
    secure_boot = optional(bool)
  })
  default     = null
  description = "Security options (e.g., secure_boot for Shielded VMs)"
}

variable "network" {
  type = object({
    machine_cidr = optional(string)
    service_cidr = optional(string)
    pod_cidr     = optional(string)
    host_prefix  = optional(number)
  })
  default     = null
  description = "Cluster network CIDRs (machine_cidr, service_cidr, pod_cidr, host_prefix)"
}

# Autoscaling (default worker pool)
variable "autoscaling" {
  type = object({
    min_replicas = number
    max_replicas = number
  })
  default     = null
  description = "Autoscaling for default worker pool (min_replicas, max_replicas)"
}

# Wait
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

# Admin
variable "create_admin" {
  type        = bool
  default     = false
  description = "Create HTPasswd IDP and cluster-admin user"
}

variable "admin_username" {
  type        = string
  default     = "admin"
  description = "Admin username (when create_admin is true)"
}

variable "admin_password" {
  type        = string
  default     = null
  sensitive   = true
  description = "Admin password (optional; auto-generated if null when create_admin is true)"
}

# Machine pools
variable "machine_pools" {
  type = list(object({
    name                = string
    instance_type       = string
    autoscaling_enabled = optional(bool, true)
    min_replicas        = optional(number)
    max_replicas        = optional(number)
    replicas            = optional(number)
    availability_zones  = optional(list(string))
    labels              = optional(map(string), {})
    taints = optional(list(object({
      key    = string
      value  = string
      effect = string
    })), [])
    root_volume_size = optional(number)
    secure_boot      = optional(bool, false)
  }))
  default     = []
  description = "Additional machine pools. If autoscaling_enabled = true, set min_replicas and max_replicas. If false, set replicas. Names 'worker' and 'workers-*' are reserved."

  validation {
    condition = alltrue([
      for mp in var.machine_pools : (
        (mp.autoscaling_enabled && mp.min_replicas != null && mp.max_replicas != null && mp.replicas == null) ||
        (!mp.autoscaling_enabled && mp.replicas != null && mp.min_replicas == null && mp.max_replicas == null)
      )
    ])
    error_message = "For each machine pool: if autoscaling_enabled = true, set min_replicas and max_replicas (replicas must be null). If false, set replicas (min_replicas and max_replicas must be null)."
  }
}
