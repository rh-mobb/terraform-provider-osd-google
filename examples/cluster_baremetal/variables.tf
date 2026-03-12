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
  default     = "my-baremetal-cluster"
  description = "Name of the cluster (must match terraform/wif_config)"
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

variable "gcp_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region"
}

variable "compute_machine_type" {
  type        = string
  default     = "c3-standard-192-metal"
  description = "Bare metal machine type for worker nodes. Use osdgoogle_machine_types data source to find available types."
}

variable "compute_nodes" {
  type        = number
  default     = 3
  description = "Number of worker nodes"
}

variable "availability_zone" {
  type        = string
  default     = "us-central1-a"
  description = "Single zone for the cluster. The machine type must be available in this zone; e.g. c3-standard-192-metal is in us-central1-a but not us-central1-b. Use google_compute_machine_types to verify."
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
