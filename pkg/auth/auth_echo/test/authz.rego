# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package authz

import future.keywords.in

hasAdminAccess {
    some role in input["realm_access/roles"] # iteration
    ["lp-admin-role"][_] == role
}

hasWriteAccess {
    some role in input["realm_access/roles"] # iteration
    ["lp-admin-role", "lp-read-write-role"][_] == role
}

hasReadAccess {
    some role in input["realm_access/roles"] # iteration
    ["lp-admin-role", "lp-read-write-role", "lp-read-only-role"][_] == role
}