// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"regexp"

	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
)

type ContextKey string // SA1029

const (
	ActiveProjectIdHeaderKey  = "Activeprojectid"
	ActiveProjectIdContextKey = ContextKey(ActiveProjectIdHeaderKey)
	DefaultProjectId          = "00000000-0000-0000-0000-000000000000"

	m2mClientRole          = "en-agent-rw"
	jwtRolesKey            = "realm_access/roles"
	roleProjectIdSeparator = "_"

	// relaxed uuid regex, replace with strict regex if required
	uuidPattern = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
)

var (
	log = logging.GetLogger("tenant")

	uuidRegex = regexp.MustCompile(uuidPattern)
)
