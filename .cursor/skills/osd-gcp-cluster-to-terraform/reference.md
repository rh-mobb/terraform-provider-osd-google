# Reference — imports, IDs, and caveats

Use this file when generating `IMPORT.md` or answering adoption questions. Prefer verifying against current provider source under `provider/*/resource.go`.

## Import ID formats (typical)

| Resource | Import ID | Notes |
|----------|-----------|--------|
| `osdgoogle_cluster` | OCM **cluster** ID | Passthrough to `id`. Required config still needs `name`, `cloud_region`, `gcp_project_id` in schema. |
| `osdgoogle_wif_config` | WIF config OCM **id** | Passthrough to `id`. |
| `osdgoogle_machine_pool` | OCM machine pool **id** | Importer sets `id` only; `cluster_id` must be present in **state** for `Read` — confirm current provider behavior before promising one-step import; composite `cluster_id/pool_id` may be needed in future. |
| `osdgoogle_cluster_admin` | `cluster_id/idp_id` | IdP id from OCM. |
| `osdgoogle_dns_domain` | Domain resource **id** | OCM id for the dns domain object. |
| `osdgoogle_cluster_waiter` | `cluster_id` | Passthrough to `cluster_id`. |

**Google provider** resources use provider-documented import formats (VPC, subnets, PSC, etc.) — link to HashiCorp registry docs in generated `IMPORT.md`.

## OCM vs GCP state

- **`osdgoogle_*`** — OpenShift Cluster Manager API; state keys are OCM-centric.
- **VPC, firewall, PSC consumer/producer, IAM bindings** — `google_*` resources; import or data sources per Terraform Google provider.

## Cluster resource honesty

- `osdgoogle_cluster` **Update** patches only a **subset** of fields (e.g. domain prefix, multi-AZ, properties). Many create-time settings are not driven by post-import drift the same way.
- **`populateState`** does not fill every optional attribute from OCM. After import, **plan** may show gaps or optional arguments not refreshed from API — document “align or omit optional blocks” for the user.

## Machine pools

- Default worker pool naming is **reserved** for `osdgoogle_machine_pool` (see provider validator / docs).
- Non-default pools are the usual TF adoption target.

## Security

- Never commit `terraform.tfvars` with keys or OCM tokens. Recommend `.gitignore` patterns and `-var-file` for local use.
