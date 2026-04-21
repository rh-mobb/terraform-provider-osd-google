variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (optional; use OSDGOOGLE_TOKEN env var)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster and bastion VM"
}

variable "cluster_name" {
  type        = string
  default     = "my-private-cluster"
  description = "Cluster name (must match terraform/wif_config display_name prefix)"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "gcp_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region for the cluster and bastion"
}

variable "bastion_zone" {
  type        = string
  default     = "us-central1-a"
  description = "GCP zone for the bastion VM (must be within gcp_region)"
}

variable "bastion_machine_type" {
  type        = string
  default     = "e2-medium"
  description = "GCP machine type for the bastion VM (e2-micro is too small for dnf/yum installs)"
}

variable "bastion_use_worker_subnet" {
  type        = bool
  default     = true
  description = "Place the bastion in the worker subnet instead of a dedicated bastion subnet. Gives the bastion identical firewall rules, NAT, and Private Google Access to worker nodes. Useful for debugging cluster connectivity."
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
