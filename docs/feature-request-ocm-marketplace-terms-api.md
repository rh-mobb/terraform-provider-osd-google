# Feature Request: Accept GCP Marketplace Terms via Standalone API

**Target:** OCM (OpenShift Cluster Manager) API  
**Component:** Authorizations / Clusters Mgmt  
**Proposed by:** tf-provider-osd-google maintainers  

---

## Summary

Provide a standalone OCM API endpoint to accept GCP Marketplace terms **without** creating a cluster. Currently, terms can only be accepted by passing `marketplace-gcp-terms=true` as a query parameter on the clusters Add (create) request.

## Problem

Accepting GCP Marketplace terms is tied to cluster creation:

- **Single point of coupling** — Terms acceptance is only possible as part of `POST /api/clusters_mgmt/v1/clusters` with `?marketplace-gcp-terms=true`.
- **Non-deterministic workflows** — Tools (e.g. Terraform, CI/CD) cannot cleanly separate "ensure terms are accepted" from "create cluster." The first cluster create may succeed or fail depending on whether terms were previously accepted via the console or another flow.
- **Orchestration complexity** — Operators cannot pre-accept terms before running cluster provisioning, making pipelines harder to reason about and debug.
- **Idempotency** — There is no way to run an idempotent "ensure terms accepted" step; the only option is to attempt cluster creation and handle the terms-related failure.

## Proposed Solution

Add an endpoint that explicitly accepts GCP Marketplace terms for the authenticated account/organization, independent of cluster creation.

### Option A: New endpoint under authorizations

```
POST /api/authorizations/v1/gcp_marketplace_terms
```

**Request body (optional):**
```json
{
  "accept": true
}
```

**Response:** Same shape as existing terms review responses, e.g.:
```json
{
  "terms_available": true,
  "terms_required": false,
  ...
}
```

- If terms were already accepted → 2xx, `terms_required: false`
- If terms were just accepted → 2xx, `terms_required: false`
- If terms cannot be accepted (e.g. invalid account) → 4xx with appropriate error

### Option B: Extend SelfTermsReview

If `self_terms_review` (or equivalent) already supports a "check" flow, extend it to support an explicit "accept" action when terms are required:

```
POST /api/authorizations/v1/self_terms_review
Content-Type: application/json

{
  "event_code": "gcp_marketplace",
  "site_code": "...",
  "action": "accept"   // or similar
}
```

## Benefits

- **Deterministic tooling** — Terraform and other tools can run "accept terms if needed" as a discrete step before cluster creation.
- ** clearer separation of concerns** — Terms acceptance (one-time, account-level) is distinct from cluster provisioning (per-cluster).
- **Better error handling** — Operators can fail fast on terms acceptance without triggering cluster creation logic.
- **Idempotent pre-requisite** — Pipelines can include an idempotent "ensure marketplace terms accepted" step that always succeeds when terms are already accepted.

## Use Case Example (Terraform)

```hcl
# Conceptual: standalone resource to accept terms
resource "osdgoogle_marketplace_terms" "gcp" {}

# Cluster creation no longer needs to guess at terms state
resource "osdgoogle_cluster" "example" {
  depends_on = [osdgoogle_marketplace_terms.gcp]
  ...
}
```

## References

- Current behavior: `marketplace-gcp-terms` query parameter on `ClustersAddRequest`
- Related: `terms_review`, `self_terms_review` in `authorizations/v1`
