# Terraform Provider for OpenShift Dedicated on Google Cloud

A [Terraform](https://www.terraform.io/) provider for managing [OpenShift Dedicated (OSD)](https://cloud.redhat.com/openshift) clusters on Google Cloud Platform (GCP). The provider uses the OpenShift Cluster Manager (OCM) API to provision and manage OSD clusters with support for Workload Identity Federation (WIF), Private Service Connect (PSC), Shared VPC, and CMEK encryption.

## Features

### Resources

| Resource | Description |
|----------|-------------|
| `osdgoogle_cluster` | Create and manage OSD clusters on GCP |
| `osdgoogle_cluster_admin` | HTPasswd identity provider with cluster-admin user |
| `osdgoogle_wif_config` | Workload Identity Federation configuration for OSD on GCP |
| `osdgoogle_machine_pool` | Machine pools for worker nodes |
| `osdgoogle_cluster_waiter` | Wait for a cluster to reach a desired state |
| `osdgoogle_dns_domain` | DNS domain reservation |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `osdgoogle_versions` | List available OpenShift versions |
| `osdgoogle_machine_types` | List GCP machine types by region |
| `osdgoogle_regions` | List available GCP regions |
| `osdgoogle_wif_config` | Look up WIF config by display name or ID |

### Supported Cluster Configuration

- **Workload Identity Federation (WIF)** – Use WIF instead of service account keys
- **Private Service Connect (PSC)** – Private connectivity to Red Hat services
- **Shared VPC** – Deploy into an existing shared VPC
- **CMEK** – Customer-managed encryption keys
- **Shielded VM (Secure Boot)** – Per-cluster or per-machine-pool
- **Autoscaling** – Min/max replicas for worker nodes

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/) 1.24+ (for building from source)
- OCM token from [console.redhat.com](https://console.redhat.com/openshift/token/rosa)
- GCP project with billing enabled and OSD entitlements

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    osdgoogle = {
      source  = "registry.terraform.io/rh-mobb/osd-google"
      version = ">= 0.0.1"
    }
  }
}
```

### From Source (for development)

See the [Development Workflow](#development-workflow) section below for how to build and test the provider locally using `dev_overrides`.

## Authentication

The provider requires credentials to access the OCM API. You can use either a **token** (recommended for interactive use) or **client credentials** (for CI/CD or automation).

### Obtaining an OCM token

1. Log in to [Red Hat Hybrid Cloud Console](https://console.redhat.com)
2. Go to **OpenShift** → **Token** ([direct link](https://console.redhat.com/openshift/token/rosa))
3. Click **Load token** or **Copy** to copy your offline token
4. Use it in the provider:

   ```bash
   export OSDGOOGLE_TOKEN="your-token-here"
   ```

   Or in Terraform:

   ```hcl
   provider "osdgoogle" {
     token = var.ocm_token
   }
   ```

Offline tokens are long-lived and can be refreshed automatically. Access tokens are short-lived (about 1 hour).

### Obtaining client credentials

Client credentials (`client_id` + `client_secret`) are used for non-interactive or programmatic access (e.g. CI/CD pipelines). They use the OAuth2 client credentials grant.

To obtain client credentials for OCM:

1. Contact your Red Hat account team or open a support case to request OAuth2 API credentials for programmatic OCM access
2. Alternatively, if your organization has set up a service account or OAuth2 client in the Red Hat SSO realm (`redhat-external`), use those credentials

Once you have them:

```bash
export OSDGOOGLE_CLIENT_ID="your-client-id"
export OSDGOOGLE_CLIENT_SECRET="your-client-secret"
```

Or in Terraform (use variables for sensitive values):

```hcl
provider "osdgoogle" {
  client_id     = var.ocm_client_id
  client_secret = var.ocm_client_secret  # mark as sensitive
}
```

For more details, see [Red Hat OCM CLI documentation](https://access.redhat.com/articles/6114701) or run `ocm login --help`.

## Provider Configuration

You can authenticate using either a **token** or **client credentials** (same options as the OCM CLI):

### Option 1: Token (offline or access token)

```hcl
provider "osdgoogle" {
  token = var.ocm_token  # or use OSDGOOGLE_TOKEN env var
}
```

### Option 2: Client ID and Client Secret (like `ocm login --client-id x --client-secret y`)

```hcl
provider "osdgoogle" {
  client_id     = var.client_id      # or use OSDGOOGLE_CLIENT_ID env var
  client_secret = var.client_secret  # or use OSDGOOGLE_CLIENT_SECRET env var
}
```

### Full configuration

```hcl
provider "osdgoogle" {
  token = var.ocm_token  # OR client_id + client_secret

  url         = "https://api.openshift.com"
  token_url   = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
  trusted_cas = file("path/to/ca.pem")
  insecure    = false
}
```

| Argument | Description |
|----------|-------------|
| `token` | OCM offline or access token (sensitive). Use with `OSDGOOGLE_TOKEN` env var. |
| `client_id` | OAuth client identifier for client credentials flow. Use with `OSDGOOGLE_CLIENT_ID` env var. |
| `client_secret` | OAuth client secret (sensitive). Use with `OSDGOOGLE_CLIENT_SECRET` env var. |
| `url` | OCM API URL (default: `https://api.openshift.com`) |
| `token_url` | OpenID token endpoint |
| `trusted_cas` | PEM CA bundle for TLS |
| `insecure` | Skip TLS verification (not for production) |

## Quick Start

```hcl
provider "osdgoogle" {
  token = var.ocm_token
}

resource "osdgoogle_cluster" "example" {
  name                 = "my-osd-cluster"
  cloud_region         = "us-central1"
  gcp_project_id       = var.gcp_project_id
  version              = "4.16.1"
  compute_nodes        = 3
  compute_machine_type = "custom-4-16384"
}

output "api_url" {
  value = osdgoogle_cluster.example.api_url
}

output "console_url" {
  value = osdgoogle_cluster.example.console_url
}
```

CCS clusters (your own GCP project) require `wif_config_id` or `gcp_authentication`. Create the WIF config in OCM using [`terraform/wif_config/`](terraform/wif_config/README.md) before provisioning a cluster — see [cluster](examples/cluster) for an example.

## Examples

The examples use the [Terraform Registry](https://registry.terraform.io/providers/rh-mobb) source by default.
For local development, use `dev_overrides` so Terraform uses your local build without changing any example files — see [Development Workflow](#development-workflow).

| Example | Description |
|---------|-------------|
| [cluster](examples/cluster) | Basic cluster with OCM managd VPC |
| [cluster_with_vpc](examples/cluster_with_vpc) | Cluster with module-managed VPC (BYOVPC) |
| [cluster_psc](examples/cluster_psc) | Cluster with Private Service Connect and Secure Boot |
| [cluster_shared_vpc](examples/cluster_shared_vpc) | Cluster using a Shared VPC |
| [cluster_multi_az](examples/cluster_multi_az) | Multi-AZ cluster with bare metal machine pool |

Every `make example.<name>` target handles the full lifecycle — WIF config ([`terraform/wif_config/`](terraform/wif_config/README.md)) is applied first, then the cluster. On destroy, the cluster is destroyed first, then the WIF config:

```bash
make build
export OSDGOOGLE_TOKEN="your-token"
gcloud auth application-default login
make example.cluster              # Create WIF + cluster
make example.cluster.destroy      # Destroy cluster + WIF
```

To test with your **local provider build** (build, install to `~/.terraform.d/plugins`, re-init, then run), use the `dev.*` targets:

```bash
make dev.cluster.apply             # Apply with local provider build
make dev.cluster.plan              # Plan with local provider build
make dev.cluster.destroy           # Destroy with local provider build
```

The Make targets infer `gcp_project_id` from `gcloud config` and `cluster_name` from your username. See [examples/cluster/README.md](examples/cluster/README.md) for details.

## AI Agent Development

When developing with an AI coding assistant (Cursor, Claude, Copilot, etc.), clone the upstream reference repositories into `references/` before starting. These repos are gitignored and provide agents with offline context for the OCM API, the Go SDK, and the canonical RHCS provider structure — reducing hallucinations and improving code quality significantly.

```bash
make references
```

Run the same command at any time to pull the latest changes from each repo's default branch. See [AGENTS.md](AGENTS.md#references) for a description of each reference and what it is useful for.

## Development

### Requirements

- Go 1.24+
- [Terraform](https://www.terraform.io/downloads)
- [jq](https://jqlang.github.io/jq/) (for `make install`)

### Development Workflow

The recommended way to develop and test the provider locally is with Terraform's `dev_overrides`.
This lets you build the provider once and use it in any example directory without changing the `required_providers` source or running `terraform init`.

#### One-time setup

1. Build the provider and get the `~/.terraformrc` snippet:

```bash
make dev-setup
```

2. Add the printed block to `~/.terraformrc` (create the file if it doesn't exist):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/rh-mobb/osd-google" = "/path/to/terraform-provider-osd-google"
  }
  direct {}
}
```

Replace the path with the actual repo directory printed by `make dev-setup`.

#### Iterative development loop

After the one-time setup, the dev cycle is:

```bash
make build
make example.cluster                # or: cd examples/cluster && terraform plan
```

No `terraform init` is needed when using `dev_overrides` — Terraform finds the provider binary directly.

#### Alternative: dev.* targets (no dev_overrides required)

If you prefer not to use `dev_overrides`, the `dev.*` targets build and install the provider to `~/.terraform.d/plugins`, clear lock files, re-init, and then run:

```bash
export OSDGOOGLE_TOKEN="your-token"
gcloud auth application-default login
make dev.cluster.apply              # Install provider + WIF + cluster
make dev.cluster.plan               # Plan only
make dev.cluster.destroy            # Destroy cluster + WIF
```

Use `dev.<example>` for any example: `dev.cluster`, `dev.cluster_baremetal`, `dev.cluster_with_vpc`, `dev.cluster_psc`, `dev.cluster_shared_vpc`, `dev.cluster_multi_az`. Override `GCP_PROJECT_ID` and `CLUSTER_NAME` as needed. Each run uses the freshly built provider.

> **Note:** Terraform prints a warning about `dev_overrides` being active. This is expected and safe to ignore during development.

When you're done developing, remove or comment out the `dev_overrides` block in `~/.terraformrc` to go back to using the registry provider.

#### Debugging with Delve

To step through provider code with a debugger:

```bash
go build -gcflags="all=-N -l" -o terraform-provider-osd-google .
dlv exec ./terraform-provider-osd-google -- -debug
```

Delve prints a `TF_REATTACH_PROVIDERS` value. Export it in another terminal:

```bash
export TF_REATTACH_PROVIDERS='<value printed by delve>'
cd examples/cluster
terraform apply    # attaches to the running provider process
```

#### Provider logging

Set `TF_LOG` to see provider-level debug output:

```bash
TF_LOG=DEBUG terraform apply
TF_LOG_PROVIDER=TRACE terraform plan   # provider logs only (no Terraform core noise)
```

### Alternative: Install to plugins directory

For local development before the provider is published, use the local plugins directory:

1. Install the provider:

```bash
make install
```

2. Terraform automatically checks `~/.terraform.d/plugins` for local providers. Ensure no `provider_installation` block in `~/.terraformrc` overrides this. If you had `dev_overrides` for the registry source, remove it.

3. Run `terraform init` in the example directory.

### Run Tests

```bash
# Unit tests (no infrastructure required)
make unit-test

# Subsystem tests (uses OCM mock server; requires make install)
make subsystem-test

# Acceptance tests (real OCM + GCP; requires OCM_TOKEN and GCP_PROJECT_ID)
make acceptance-test
```

### Generate Documentation

```bash
make tools   # install tfplugindocs
make docs    # generate docs in docs/ from templates and schema
```

Run `make docs` before every PR when you change provider schema, resources, data sources, or templates. CI fails if `docs/` is out of date.

### Code Formatting

```bash
make fmt
```

## Documentation

Provider documentation is generated in the [docs/](docs/) directory:

- [Provider configuration](docs/index.md)
- [Resources](docs/resources/)
- [Data sources](docs/data-sources/)

## Project Structure

```
.
├── main.go                 # Provider entry point
├── provider/
│   ├── provider.go         # Provider schema and configuration
│   ├── cluster/            # osdgoogle_cluster
│   ├── wif_config/         # osdgoogle_wif_config
│   ├── machine_pool/       # osdgoogle_machine_pool
│   ├── cluster_waiter/     # osdgoogle_cluster_waiter
│   ├── dns_domain/         # osdgoogle_dns_domain
│   ├── datasources/        # versions, machine_types, regions
│   └── common/             # Shared helpers
├── terraform/              # Shared Terraform configs (applied before examples)
│   └── wif_config/         # WIF config in OCM ([README](terraform/wif_config/README.md) explains why separate apply)
├── subsystem/              # OCM mock integration tests
├── acceptance/             # Real API acceptance tests
├── examples/               # Example configurations
└── docs/                   # Generated documentation
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code style, and how to submit changes. By participating, you agree to our [Code of Conduct](CODE_OF_CONDUCT.md).

## License

Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full text.
