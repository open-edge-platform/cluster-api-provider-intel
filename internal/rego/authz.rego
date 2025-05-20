# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package authz

import future.keywords.in

hasWriteAccess := true if {
    some role in input["realm_access/roles"] # iteration

    # Check if the request has the '<ProjectId>_en-agent-rw' permission
    regex.match("^(([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}_)en-agent-rw)", role)
}

hasReadAccess := true if {
    some role in input["realm_access/roles"] # iteration

    # Check if the request has the '<ProjectId>_en-agent-rw' permission
    regex.match("^(([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}_)en-agent-rw)", role)
}

# TODO: This will be removed in subsequent versions.  Still needed for grpc_server tests to pass.
hasWriteAccess := true if {
    some role in input["realm_access/roles"] # iteration
    # We expect:
    # - with MT: [PROJECT_UUID]_node-agent-readwrite-role
    # - without MT: node-agent-readwrite-role
    regex.match("^(([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}_)?node-agent-readwrite-role)", role)
}

hasReadAccess := true if {
    some role in input["realm_access/roles"] # iteration
    # We expect:
    # - with MT: [PROJECT_UUID]_node-agent-readwrite-role
    # - without MT: node-agent-readwrite-role
    regex.match("^(([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}_)?node-agent-readwrite-role)", role)
}
