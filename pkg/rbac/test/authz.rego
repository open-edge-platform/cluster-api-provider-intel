# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package authz

import future.keywords.in

hasWriteAccess := true if {
    some role in input["realm_access/roles"] # iteration
    ["clusters-write-role"][_] == role
}

hasReadAccess := true if {
    some role in input["realm_access/roles"] # iteration
    ["clusters-read-role"][_] == role
}