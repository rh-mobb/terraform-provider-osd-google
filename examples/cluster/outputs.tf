output "cluster_id" {
  value       = module.cluster.cluster_id
  description = "OCM cluster ID"
}

output "wif_config_id" {
  value       = module.cluster.wif_config_id
  description = "WIF config ID (from data source)"
}

output "api_url" {
  value       = module.cluster.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = module.cluster.console_url
  description = "OpenShift web console URL"
}

output "admin_username" {
  value       = module.cluster.admin_username
  description = "Cluster admin username (use with oc login)"
}

output "admin_password" {
  value       = module.cluster.admin_password
  description = "Cluster admin password (sensitive; omit from logs)"
  sensitive   = true
}
