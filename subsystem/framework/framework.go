/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	"github.com/onsi/gomega/ghttp"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

// All tests will use this API TestServer and Terraform runner.
var (
	TestServer *ghttp.Server
	Terraform  *TerraformRunner
	TestingT   *testing.T
)

// TerraformRunnerBuilder contains the data and logic needed to build a terraform runner.
type TerraformRunnerBuilder struct {
	url   string
	ca    string
	token string
}

// RunOutput captures the output of a terraform command.
type RunOutput struct {
	Out      string
	Err      string
	ExitCode int
}

// VerifyErrorContainsSubstring checks that the error output contains the given substring.
func (ro *RunOutput) VerifyErrorContainsSubstring(sub string) {
	Expect(ro.Err).To(ContainSubstring(sub))
}

// TerraformRunner contains the data and logic needed to run Terraform.
type TerraformRunner struct {
	binary string
	dir    string
	env    []string
}

// NewTerraformRunner creates a new Terraform runner builder.
func NewTerraformRunner() *TerraformRunnerBuilder {
	return &TerraformRunnerBuilder{}
}

// URL sets the URL of the OCM API server.
func (b *TerraformRunnerBuilder) URL(value string) *TerraformRunnerBuilder {
	b.url = value
	return b
}

// CA sets the trusted certificates used to connect to the OCM API server.
func (b *TerraformRunnerBuilder) CA(value string) *TerraformRunnerBuilder {
	b.ca = value
	return b
}

// Token sets the authentication token used to connect to the OCM API server.
func (b *TerraformRunnerBuilder) Token(value string) *TerraformRunnerBuilder {
	b.token = value
	return b
}

// Build uses the information stored in the builder to create a new Terraform runner.
func (b *TerraformRunnerBuilder) Build() *TerraformRunner {
	ExpectWithOffset(1, b.url).ToNot(BeEmpty())
	ExpectWithOffset(1, b.ca).ToNot(BeEmpty())
	ExpectWithOffset(1, b.token).ToNot(BeEmpty())

	tfBinary, err := exec.LookPath("terraform")
	Expect(err).ToNot(HaveOccurred())

	tmpDir, err := os.MkdirTemp("", "osdgoogle-test-*.d")
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	mainPath := filepath.Join(tmpDir, "main.tf")
	mainContent := EvaluateTemplate(`
		terraform {
		  required_providers {
		    osdgoogle = {
		      source = "terraform.local/local/osd-google"
		      version = ">= 0.0.1"
		    }
		  }
		}

		provider "osdgoogle" {
		  url         = "{{ .URL }}"
		  token       = "{{ .Token }}"
		  trusted_cas = file("{{ .CA }}")
		}
		`,
		"URL", b.url,
		"Token", b.token,
		"CA", strings.ReplaceAll(b.ca, "\\", "/"),
	)
	err = os.WriteFile(mainPath, []byte(mainContent), 0600)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	envMap := map[string]string{}
	for _, text := range os.Environ() {
		idx := strings.Index(text, "=")
		if idx > 0 {
			envMap[text[0:idx]] = text[idx+1:]
		} else {
			envMap[text] = ""
		}
	}
	envMap["IS_TEST"] = "true"
	envMap["TF_LOG"] = "DEBUG"

	envList := make([]string, 0, len(envMap))
	for name, value := range envMap {
		envList = append(envList, name+"="+value)
	}

	initCmd := exec.Command(tfBinary, "init")
	initCmd.Env = envList
	initCmd.Dir = tmpDir
	initCmd.Stdout = GinkgoWriter
	initCmd.Stderr = GinkgoWriter
	err = initCmd.Run()
	if err != nil {
		code := 1
		if initCmd.ProcessState != nil {
			code = initCmd.ProcessState.ExitCode()
		}
		Fail(fmt.Sprintf("Terraform init finished with exit code %d", code), 1)
	}

	return &TerraformRunner{
		binary: tfBinary,
		dir:    tmpDir,
		env:    envList,
	}
}

// Source sets the Terraform source of the test.
func (r *TerraformRunner) Source(text string) {
	file := filepath.Join(r.dir, "test.tf")
	err := os.WriteFile(file, []byte(text), 0600)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

// Run runs a terraform command.
func (r *TerraformRunner) Run(args ...string) RunOutput {
	cmd := exec.Command(r.binary, args...)
	var outb, errb bytes.Buffer
	cmd.Env = r.env
	cmd.Dir = r.dir
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	if err != nil && code == 0 {
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
	}

	GinkgoWriter.Println(&outb)
	GinkgoWriter.Println(&errb)

	return RunOutput{
		Out:      outb.String(),
		Err:      errb.String(),
		ExitCode: code,
	}
}

// Validate runs the validate command.
func (r *TerraformRunner) Validate() RunOutput {
	return r.Run("validate")
}

// Apply runs the apply command.
func (r *TerraformRunner) Apply() RunOutput {
	return r.Run("apply", "-auto-approve")
}

// Destroy runs the destroy command.
func (r *TerraformRunner) Destroy() RunOutput {
	return r.Run("destroy", "-auto-approve")
}

// Import runs the import command.
func (r *TerraformRunner) Import(args ...string) RunOutput {
	return r.Run(append([]string{"import"}, args...)...)
}

// State reads the Terraform state and returns it as a parsed JSON document.
func (r *TerraformRunner) State() interface{} {
	path := filepath.Join(r.dir, "terraform.tfstate")
	_, err := os.Stat(path)
	var result interface{}
	if err == nil {
		data, err := os.ReadFile(path)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		err = json.Unmarshal(data, &result)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
	}
	return result
}

// Resource returns the resource stored in the state with the given type and identifier.
func (r *TerraformRunner) Resource(typ, name string) interface{} {
	state := r.State()
	filter := fmt.Sprintf(
		`.resources[] | select(.type == "%s" and .name == "%s") | .instances[]`,
		typ, name,
	)
	results, err := JQ(filter, state)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, results).To(
		HaveLen(1),
		"Expected exactly one resource with type '%s' and name '%s', but found %d",
		typ, name, len(results),
	)
	return results[0]
}

// Close releases all resources used by the Terraform runner.
func (r *TerraformRunner) Close() {
	_ = os.RemoveAll(r.dir)
}
