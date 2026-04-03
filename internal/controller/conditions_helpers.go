// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
)

// Helper functions for v1beta2 conditions API

// markConditionTrue sets a condition to True status.
func markConditionTrue(obj conditions.Setter, conditionType string) {
	conditions.Set(obj, metav1.Condition{
		Type:   conditionType,
		Status: metav1.ConditionTrue,
		Reason: "Ready",
	})
}

// markConditionFalse sets a condition to False status with a reason and message.
func markConditionFalse(obj conditions.Setter, conditionType string, reason string, severity clusterv1.ConditionSeverity, messageFormat string, messageArgs ...interface{}) {
	conditions.Set(obj, metav1.Condition{
		Type:    conditionType,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	})
}

// markConditionUnknown sets a condition to Unknown status with a reason and message.
func markConditionUnknown(obj conditions.Setter, conditionType string, reason string, messageFormat string, messageArgs ...interface{}) {
	conditions.Set(obj, metav1.Condition{
		Type:    conditionType,
		Status:  metav1.ConditionUnknown,
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	})
}
