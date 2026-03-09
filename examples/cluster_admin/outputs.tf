output "cluster_id" {
  value       = osdgoogle_cluster.example.id
  description = "OCM cluster ID"
}

output "api_url" {
  value       = osdgoogle_cluster.example.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.example.console_url
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
