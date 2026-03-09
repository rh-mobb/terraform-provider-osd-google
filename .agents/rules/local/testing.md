---
description: "Ginkgo/Gomega and subsystem test patterns aligned with Red Hat RHCS conventions"
globs: ["**/*_test.go", "**/subsystem/**"]
alwaysApply: false
---

# Testing Conventions (RHCS-Aligned)

Use Ginkgo and Gomega for all tests. Do not use standard `go test` with `testify`. Align with Red Hat's terraform-provider-rhcs testing patterns.

## Test Framework

- **Ginkgo/Gomega** for unit and subsystem tests
- Run tests with `ginkgo run -r` (not `go test ./...`)
- Use `make unit-test` and `make subsystem-test`

## Suite Setup

```go
func TestSubsystem(t *testing.T) {
    RegisterFailHandler(Fail)
    TestingT = t
    RunSpecs(t, "Subsystem Suite Name")
}

var _ = BeforeEach(func() {
    format.MaxLength = 0  // full diff output on failure
    TestServer, ca = MakeTCPTLSServer()
    token := MakeTokenString("Bearer", 10*time.Minute)
    Terraform = NewTerraformRunner().
        URL(TestServer.URL()).
        CA(ca).
        Token(token).
        Build()
})

var _ = AfterEach(func() {
    TestServer.Close()
    Terraform.Close()
})
```

- Import `ocm-sdk-go/testing` for `MakeTCPTLSServer`, `MakeTokenString`, `EvaluateTemplate`, `CombineHandlers`
- Import `ghttp` for mock HTTP handlers

## Subsystem Test Flow

1. **Mock API**: `TestServer.AppendHandlers(CombineHandlers(...))` with `VerifyRequest`, `RespondWithJSON`, etc.
2. **Write HCL**: `Terraform.Source(\`resource "..." "..." { ... }\`)`
3. **Apply/Destroy**: `runOutput := Terraform.Apply()` or `Terraform.Destroy()`
4. **Assert exit code**: `Expect(runOutput.ExitCode).To(BeZero())`
5. **Assert state**: `Terraform.Resource("type", "name")` then `Expect(resource).To(MatchJQ(path, value))`

## Unit Tests in Provider Packages

- Use Ginkgo `Describe`/`It` blocks
- Mock OCM clients with `mockgen`-generated mocks (gomock)
- Test CRUD logic, validation, and state population in isolation

## Init Test

Include an init test that validates provider configuration:

```go
var _ = Describe("Init", func() {
    It("Downloads and installs the provider", func() {
        Expect(Terraform.Validate()).To(BeZero())
    })
})
```

## Subsystem Test Prerequisites

- `make subsystem-test` depends on `make install` — provider must be installed to `~/.terraform.d/plugins/` before running subsystem tests
