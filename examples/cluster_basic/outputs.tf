output "osd_ccs_admin_email" {
  value       = data.google_service_account.osd_ccs_admin.email
  description = "OSD CCS admin service account email"
}

output "cluster_id" {
  value       = osdgoogle_cluster.example.id
  description = "OCM cluster ID"
}

output "cluster_state" {
  value       = osdgoogle_cluster.example.state
  description = "Cluster state"
}

output "api_url" {
  value       = osdgoogle_cluster.example.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.example.console_url
  description = "OpenShift web console URL"
}
