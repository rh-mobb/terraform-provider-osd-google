# OSD cluster with Workload Identity Federation (WIF) + Customer-Managed Encryption Key (CMEK)
#
# Tests whether OCM supports provisioning a WIF-authenticated CCS cluster with CMEK.
# The KMS key ring and crypto key are created here; the WIF config must already exist
# (create it with terraform/wif_config/ or make example.cluster first).
#
# The Compute Engine Service Agent is granted the KMS encrypter/decrypter role so that
# GCP can use the customer-managed key when creating encrypted worker node disks.
#
# Usage: make example.wif-kms-key

data "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
}

data "google_project" "project" {
  project_id = var.gcp_project_id
}

# Provision WIF IAM bindings in GCP (workload identity pool, service accounts, roles)
module "wif_gcp" {
  source = "../../modules/osd-wif-gcp"

  project_id   = var.gcp_project_id
  display_name = data.osdgoogle_wif_config.wif.display_name
  pool_id      = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.pool_id
  identity_provider = {
    identity_provider_id = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.identity_provider_id
    issuer_url           = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.issuer_url
    jwks                 = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.jwks
    allowed_audiences    = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.allowed_audiences
  }
  service_accounts         = data.osdgoogle_wif_config.wif.gcp.service_accounts
  support                  = data.osdgoogle_wif_config.wif.gcp.support
  impersonator_email       = data.osdgoogle_wif_config.wif.gcp.impersonator_email
  federated_project_id     = try(data.osdgoogle_wif_config.wif.gcp.federated_project_id, null) != "" ? try(data.osdgoogle_wif_config.wif.gcp.federated_project_id, null) : null
  federated_project_number = try(data.osdgoogle_wif_config.wif.gcp.federated_project_number, "") != "" ? data.osdgoogle_wif_config.wif.gcp.federated_project_number : tostring(data.google_project.project.number)
}

# KMS key ring scoped to the cluster region
resource "google_kms_key_ring" "osd" {
  name     = "${var.cluster_name}-keyring"
  location = var.cloud_region
}

# Customer-managed encryption key for worker node disks
resource "google_kms_crypto_key" "osd" {
  name     = "${var.cluster_name}-key"
  key_ring = google_kms_key_ring.osd.id
  purpose  = "ENCRYPT_DECRYPT"
}

# Dedicated service account for KMS access.
# OCM validates this SA exists in GCP before cluster creation, so we create it
# explicitly rather than relying on the lazily-provisioned Compute Engine Service Agent.
resource "google_service_account" "kms" {
  account_id   = "${var.cluster_name}-kms"
  display_name = "CMEK access for OSD cluster ${var.cluster_name}"
  project      = var.gcp_project_id
}

# Grant the KMS SA encrypter/decrypter access on the crypto key.
# OCM validates this SA exists and passes it to GCP as the disk encryption key SA.
resource "google_kms_crypto_key_iam_member" "kms_sa" {
  crypto_key_id = google_kms_crypto_key.osd.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${google_service_account.kms.email}"
}

# Grant the Compute Engine Service Agent encrypter/decrypter access.
# OCM independently validates this agent has the role — it is the GCP-managed
# identity that performs the actual disk encryption when launching worker nodes.
# This SA is created lazily by GCP; it exists once the Compute Engine API has
# been invoked in the project (e.g. during WIF IAM setup above).
resource "google_kms_crypto_key_iam_member" "compute_agent" {
  crypto_key_id = google_kms_crypto_key.osd.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.project.number}@compute-system.iam.gserviceaccount.com"
}

# OSD cluster: WIF authentication + CMEK disk encryption
resource "osdgoogle_cluster" "cluster" {
  depends_on = [module.wif_gcp, google_kms_crypto_key_iam_member.kms_sa, google_kms_crypto_key_iam_member.compute_agent]

  name           = var.cluster_name
  cloud_region   = var.cloud_region
  gcp_project_id = var.gcp_project_id
  wif_config_id  = data.osdgoogle_wif_config.wif.id
  version        = var.openshift_version
  compute_nodes  = var.compute_nodes
  ccs_enabled    = true

  gcp_encryption_key = {
    kms_key_service_account = google_service_account.kms.email
    key_location            = var.cloud_region
    key_name                = google_kms_crypto_key.osd.name
    key_ring                = google_kms_key_ring.osd.name
  }

  wait_for_create_complete = var.wait_for_create_complete
  wait_timeout             = var.wait_timeout
}
