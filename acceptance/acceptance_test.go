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

// Package acceptance contains acceptance tests that run against a real OCM API
// and GCP project. These tests are not run by default.
//
// Run with: ginkgo run -r ./acceptance --label-filter "Day1"
//
// Labels:
//   - Day1: Create resources (cluster, WIF config, etc.)
//   - Day2: Modify/update resources
//   - Destroy: Teardown and cleanup
//   - RequiresEnv: Needs OCM token and GCP project (skip without)
package acceptance

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance")
}

// SkipUnlessEnv skips the spec if the required env var is not set.
func SkipUnlessEnv(key string) {
	if os.Getenv(key) == "" {
		Skip("Skipping acceptance test: " + key + " not set")
	}
}
