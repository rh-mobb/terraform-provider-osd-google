# Provider TODO

Tracked improvements and feature gaps for the OSD Google Terraform provider.

## Machine type validation

The provider currently accepts any `instance_type` / `compute_machine_type` string and sends it to the OCM API without pre-validation. If the type is unavailable in the target region or zone, the error surfaces late (from OCM or GCP, not from `terraform plan`).

### Problem

- GCP bare metal types (e.g., `c3-standard-192-metal`) are only available in a subset of zones within a region.
- Some machine types may be available in GCP but not allowlisted by OCM for OSD clusters.
- Users get opaque API errors instead of actionable plan-time feedback.

### Proposed improvements

#### 1. GCP-side validation via `google_compute_machine_types`

Use the Google provider's `google_compute_machine_types` data source to verify zone-level availability before cluster or machine pool creation.

```hcl
data "google_compute_machine_types" "zone_a" {
  zone = "us-central1-a"
}

locals {
  baremetal_available = anytrue([
    for mt in data.google_compute_machine_types.zone_a.machine_types :
    mt.name == "c3-standard-192-metal"
  ])
}
```

This confirms GCP has the type in that zone. It does not confirm OCM allows it.

#### 2. OCM-side validation via `osdgoogle_machine_types`

The `osdgoogle_machine_types` data source queries the OCM GCP inquiries API for machine types allowed by OSD in a region. Enhancements needed:

- **Add `availability_zones` input** -- The OCM `CloudProviderData` schema accepts an `availability_zones` array. Passing it would let the API filter machine types by zone (if supported server-side). The provider currently only sends `gcp.project_id` and `region`.
- **Add a search/filter attribute** -- Allow users to specify a machine type name and get back a boolean or filtered list, avoiding client-side HCL filtering.
- **Return richer fields** -- The OCM `MachineType` object may include fields like `cpu`, `memory`, and `category` that the data source currently ignores (only `id` and `name` are returned).

#### 3. Provider-level plan-time validation

Add custom validators on `instance_type` (machine pool) and `compute_machine_type` (cluster) that call the OCM inquiry API during plan and produce a diagnostic if the type is not in the allowed list. This gives users immediate feedback at `terraform plan` time rather than a failing `terraform apply`.

### Considerations

- The OCM machine types inquiry requires GCP service account credentials (`gcp.client_email`) in the request body for non-WIF clusters. WIF clusters may need a different approach.
- Per-zone filtering via OCM may not be implemented server-side for GCP (the `availability_zones` field originates from AWS). This needs testing.
- Bare metal instance types do not support Secure Boot (Shielded VMs). The provider should warn or error if `security.secure_boot = true` is set alongside a bare metal `compute_machine_type` or `instance_type`.

## Default machine pool adoption

The `osdgoogle_machine_pool` resource does not adopt the cluster's default worker pool. Creating a machine pool with a reserved name will fail validation: names `worker` and `workers-*` are rejected. Users must `terraform import` the default pool to manage it.

### Proposed improvements

- Implement "magic import" logic (as in terraform-provider-rhcs): when `name = "worker"`, read the existing pool from the API instead of creating, then update it to match the plan. Until then, reserved names are rejected.
- Fix `ImportState` to accept a composite ID (`cluster_id,machine_pool_id`) so both `cluster_id` and `id` are populated on import. The current `ImportStatePassthroughID` only sets `id`.

## Multi-AZ validation

The cluster resource validates `availability_zones` when specified:

- Single-AZ (`multi_az = false`): exactly 1 zone required.
- Multi-AZ (`multi_az = true`): exactly 3 zones required.

Remaining improvements:

- `compute_nodes` should be a multiple of 3 for multi-AZ clusters.
- `multi_az` should not be updatable after creation (add `RequiresReplace` plan modifier).

## Data source enhancements

### `osdgoogle_versions`

- Add optional `search` attribute to expose OCM's server-side search (e.g., filter by channel group, major/minor version).
- Add fields like `channel_group`, `rosa_enabled`, `raw_id` from the OCM `Version` object.

### `osdgoogle_regions`

- Add optional `multi_az_only` filter to return only regions that support multiple availability zones (the `CloudRegion` object has a `supports_multi_az` field).
