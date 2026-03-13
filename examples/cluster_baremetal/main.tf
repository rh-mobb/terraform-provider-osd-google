# OSD cluster with bare metal as default compute instance type (single AZ)
#
# Uses bare metal (e.g. c3-standard-192-metal) for the default worker nodes.
# Single-AZ cluster. OCM supports multi-zone for bare metal, but the machine
# type must be available in each zone; c3-standard-192-metal is only in
# us-central1-a (not us-central1-b). Specify zones where the type exists.
#
# NOTE: Secure Boot (Shielded VMs) is NOT supported on bare metal instance types.
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)
#
# Usage: make example.cluster_baremetal
# Both configs use cluster_name; display_name is "${cluster_name}-wif".

data "osdgoogle_machine_types" "baremetal" {
  region         = var.gcp_region
  gcp_project_id = var.gcp_project_id
}

locals {
  machine_type_ids = [for item in data.osdgoogle_machine_types.baremetal.items : item.id]
}

check "machine_type_available" {
  assert {
    condition     = contains(local.machine_type_ids, var.compute_machine_type)
    error_message = <<-EOT
      Instance type '${var.compute_machine_type}' is not available in region '${var.gcp_region}'.
      Available types: ${join(", ", local.machine_type_ids)}
      For bare metal, specify availability_zones with zones where the machine type exists (e.g. us-central1-a; us-central1-b does NOT support c3-standard-192-metal).
    EOT
  }
}

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version    = var.openshift_version
  compute_nodes        = var.compute_nodes
  compute_machine_type = var.compute_machine_type
  availability_zones   = [var.availability_zone]
  ccs_enabled          = true

  create_admin   = true
  admin_password = var.admin_password != "" ? var.admin_password : null

  machine_pools = var.machine_pools
}
