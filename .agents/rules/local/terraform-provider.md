---
description: "Terraform provider patterns aligned with Red Hat RHCS conventions"
globs: ["**/*.go"]
alwaysApply: false
---

# Terraform Provider Conventions (RHCS-Aligned)

Align with Red Hat's terraform-provider-rhcs patterns for seamless handoff.

## Resource Structure

- One package per resource under `provider/<resource_name>/`
- Files: `resource.go` (CRUD handlers), `*_state.go` (state struct)
- Data sources: `*_data_source.go` or `*_datasource.go`

## Resource Implementation

```go
// resource.go
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func New() resource.Resource {
    return &Resource{}
}
```

- `Metadata`: `resp.TypeName = req.ProviderTypeName + "_resource_name"`
- `Schema`: Use Terraform Plugin Framework types (`types.String`, `types.Bool`, `types.Int64`) with plan modifiers (`stringplanmodifier.UseStateForUnknown()` where appropriate)

## State Struct

```go
// *_state.go
type ResourceState struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}
```

- Use `tfsdk` tags matching attribute names
- Use `types.String`, `types.Int64`, `types.Bool`, `types.Map`, `types.List` (never raw Go types)

## Configure

Cast provider data to `*sdk.Connection`:

```go
connection, ok := req.ProviderData.(*sdk.Connection)
if !ok {
    resp.Diagnostics.AddError("Unexpected Resource Configure Type", ...)
    return
}
r.collection = connection.ClustersMgmt().V1().<Client>()
```

## Error Handling

- Terraform errors: `resp.Diagnostics.AddError(summary, detail)` then `return`
- 404 / not found: `resp.State.RemoveResource(ctx)` with `tflog.Warn(ctx, "resource not found, removing from state", ...)`
- OCM API errors: Use `common.HandleErr(response.Error(), err)` where applicable
- Never swallow errors; always add diagnostics and return

## Immutable Attributes

For attributes that cannot change after creation, use `common.ValidateStateAndPlanEquals` (or equivalent) in Update to fail fast with a clear message.

## Update via Patch

Use `common.ShouldPatchXxx(state, plan)` to decide what to send to the API:

```go
patchLabels, ok := common.ShouldPatchMap(state.Labels, plan.Labels)
if ok {
    // build and send patch
}
```

## Shared Utilities

- Keep helpers in `provider/common/`: `helpers.go`, `tfconversions.go`, validators, plan modifiers
- Use `OptionalString`, `OptionalInt64`, `HasValue`, `StringListToArray`, `ConvertStringMapToMapType` for TF type conversions

## Logging

Use `tflog.Debug`, `tflog.Info`, `tflog.Warn`, `tflog.Error` with context — never `fmt.Println` or `log.*`.

## Code Generation

- `go:generate mockgen -source=<file> -package=<pkg> -destination=<out>` for interfaces
- `go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate ...` for docs
- Run `make generate` and ensure generated files are committed
