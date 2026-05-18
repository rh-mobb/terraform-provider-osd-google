output "cluster_id" {
  value       = osdgoogle_cluster.cluster.id
  description = "OCM cluster ID"
}

output "wif_config_id" {
  value       = data.osdgoogle_wif_config.wif.id
  description = "WIF config ID"
}

output "kms_key_ring" {
  value       = google_kms_key_ring.osd.name
  description = "KMS key ring name"
}

output "kms_crypto_key" {
  value       = google_kms_crypto_key.osd.name
  description = "KMS crypto key name"
}

output "kms_sa_email" {
  value       = google_service_account.kms.email
  description = "Service account used for KMS disk encryption"
}

output "api_url" {
  value       = osdgoogle_cluster.cluster.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.cluster.console_url
  description = "OpenShift web console URL"
}
