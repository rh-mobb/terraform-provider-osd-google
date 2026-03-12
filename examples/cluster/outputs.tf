output "cluster_id" {
  value       = osdgoogle_cluster.cluster.id
  description = "OCM cluster ID"
}

output "wif_config_id" {
  value       = data.osdgoogle_wif_config.wif.id
  description = "WIF config ID (from data source)"
}

output "api_url" {
  value       = osdgoogle_cluster.cluster.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.cluster.console_url
  description = "OpenShift web console URL"
}

output "admin_username" {
  value       = osdgoogle_cluster_admin.admin.username
  description = "Cluster admin username (use with oc login)"
}

output "admin_password" {
  value       = osdgoogle_cluster_admin.admin.password
  description = "Cluster admin password (sensitive; omit from logs)"
  sensitive   = true
}
