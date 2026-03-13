output "cluster_id" {
  value       = osdgoogle_cluster.cluster.id
  description = "OCM cluster ID"
}

output "api_url" {
  value       = osdgoogle_cluster.cluster.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.cluster.console_url
  description = "OpenShift web console URL"
}

output "domain" {
  value       = osdgoogle_cluster.cluster.domain
  description = "Cluster base domain"
}

output "state" {
  value       = osdgoogle_cluster.cluster.state
  description = "Cluster state"
}

output "infra_id" {
  value       = osdgoogle_cluster.cluster.infra_id
  description = "Infrastructure ID"
}

output "wif_config_id" {
  value       = data.osdgoogle_wif_config.wif.id
  description = "WIF config ID (from data source)"
}

output "admin_username" {
  value       = var.create_admin ? osdgoogle_cluster_admin.admin[0].username : null
  description = "Cluster admin username (when create_admin = true)"
}

output "admin_password" {
  value       = var.create_admin ? osdgoogle_cluster_admin.admin[0].password : null
  description = "Cluster admin password (sensitive; when create_admin = true)"
  sensitive   = true
}
