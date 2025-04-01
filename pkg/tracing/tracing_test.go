// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Only needed to suppress the error
	flag.String(
		"policyBundle",
		"/rego/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()

	run := m.Run() // run all tests
	os.Exit(run)
}
