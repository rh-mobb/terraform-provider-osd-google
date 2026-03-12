variable "gcp_project_id" {
  type        = string
  description = "GCP project ID"
}

variable "cluster_name" {
  type        = string
  default     = "my-vpc-cluster"
  description = "Cluster name (must match terraform/wif_config)"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (optional; use OSDGOOGLE_TOKEN env var)"
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
  description = "Additional machine pools. If autoscaling_enabled = true, set min_replicas and max_replicas. If false, set replicas."

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
