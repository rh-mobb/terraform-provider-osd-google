data "google_service_account" "osd_ccs_admin" {
  account_id = "osd-ccs-admin"
}

resource "google_service_account_key" "osd_ccs" {
  service_account_id = data.google_service_account.osd_ccs_admin.name
}

locals {
  sa_key_json = jsondecode(base64decode(google_service_account_key.osd_ccs.private_key))
}

resource "osdgoogle_cluster" "example" {
  name                = var.cluster_name
  cloud_region        = "us-central1"
  gcp_project_id      = var.gcp_project_id
  compute_nodes       = 3
  compute_machine_type = "custom-4-16384"
  ccs_enabled         = true
  wait_for_create_complete = true

  gcp_authentication = {
    client_email   = local.sa_key_json.client_email
    client_id      = local.sa_key_json.client_id
    private_key_id = local.sa_key_json.private_key_id
    private_key    = local.sa_key_json.private_key
  }
}

resource "osdgoogle_cluster_admin" "admin" {
  cluster_id = osdgoogle_cluster.example.id
  username   = "admin"
  password   = var.admin_password != "" ? var.admin_password : null
}
