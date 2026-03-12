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

variable "machine_pools" {
  type = list(object({
    name                = string
    instance_type       = string
    autoscaling_enabled = optional(bool, true)
    min_replicas        = optional(number) # When autoscaling_enabled = true
    max_replicas        = optional(number) # When autoscaling_enabled = true
    replicas            = optional(number) # When autoscaling_enabled = false
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
  description = "Additional machine pools beyond the default worker pool. If autoscaling_enabled = true, set min_replicas and max_replicas. If false, set replicas. Names 'worker' and 'workers-*' are reserved."

  validation {
    condition = alltrue([
      for mp in var.machine_pools : (
        (mp.autoscaling_enabled && mp.min_replicas != null && mp.max_replicas != null && mp.replicas == null) ||
        (!mp.autoscaling_enabled && mp.replicas != null && mp.min_replicas == null && mp.max_replicas == null)
      )
    ])
    error_message = "For each machine pool: if autoscaling_enabled = true, set min_replicas and max_replicas (replicas must be null). If autoscaling_enabled = false, set replicas (min_replicas and max_replicas must be null)."
  }
}
