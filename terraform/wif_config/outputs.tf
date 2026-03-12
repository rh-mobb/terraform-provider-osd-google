output "wif_config_id" {
  value       = osdgoogle_wif_config.wif.id
  description = "OCM WIF config ID"
}

output "wif_display_name" {
  value       = osdgoogle_wif_config.wif.display_name
  description = "WIF config display name (used by example data sources)"
}
