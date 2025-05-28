// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

// Conditions and condition Reasons for the IntelMachine object.

const (
	// HostProvisionedCondition documents the status of the provisioning of the host in Inventory.
	HostProvisionedCondition clusterv1.ConditionType = "HostProvisioned"

	// WaitingForClusterInfrastructureReason (Severity=Info) documents an IntelMachine waiting for the cluster
	// infrastructure to be ready before starting to provision the host.
	WaitingForClusterInfrastructureReason = "WaitingForClusterInfrastructure"

	// WaitingForBootstrapDataReason (Severity=Info) documents an IntelMachine waiting for the bootstrap
	// script to be ready before starting to provision the host.
	WaitingForBootstrapDataReason = "WaitingForBootstrapData"

	// WaitingforMachineBindingReason (Severity=Warning) documents an IntelMachine waiting for a valid
	// IntelMachineBinding that matches the cluster name and machine template name of the IntelMachine.
	WaitingForMachineBindingReason = "WaitingForMachineBinding"

	// HostProvisioningFailedReason (Severity=Warning) documents an IntelMachine controller detecting
	// an error while provisioning the host.  These kinds of errors are usually transient and failed
	// provisionings are automatically re-tried by the controller.
	HostProvisioningFailedReason = "HostProvisioningFailed"

	// HostDeletedReason (Severity=Error) documents an IntelMachine controller detecting
	// the underlying host has been deleted unexpectedly.
	HostDeletedReason = "HostDeleted"
)

const (
	// BootstrapExecSucceededCondition provides an observation of the IntelMachine bootstrap process.
	// 	It is set based on successful execution of bootstrap commands and on the existence of
	//	the /run/cluster-api/bootstrap-success.complete file.
	// The condition gets generated after HostProvisionedCondition is True.
	//
	// NOTE: as a difference from other providers, host bootstrap is managed by the Intel Provider's
	// Southbound Handler and Cluster Agent (not by cloud-init).
	BootstrapExecSucceededCondition clusterv1.ConditionType = "BootstrapExecSucceeded"

	// BootstrappingReason documents (Severity=Info) an IntelMachine currently executing the bootstrap
	// script that creates the Kubernetes node on the newly provisioned machine infrastructure.
	BootstrappingReason = "Bootstrapping"

	// BootstrapFailedReason documents (Severity=Warning) an IntelMachine controller detecting an error while
	// bootstrapping the Kubernetes node on the machine just provisioned; those kind of errors are usually
	// transient and failed bootstrap are automatically re-tried by the controller.
	BootstrapFailedReason = "BootstrapFailed"

	// BootstrapWaitingReason documents (Severity=Info) an IntelMachine waiting for the bootstrap script to be
	// executed on the machine.
	BootstrapWaitingReason = "BootstrapWaiting"
)

const (
	// ConnectionAliveCondition reports on whether the connection to the IntelCluster is alive.
	ConnectionAliveCondition clusterv1.ConditionType = "ConnectionAlive"
	// ConnectionNotAliveReason (Severity=Error) refers to a IntelCluster which is not alive and the connection to it is unhealthy.
	ConnectionNotAliveReason = "NoConnectionToCluster"
)

const (
	// ControlPlaneEndpointReadyCondition reports on whether a control plane endpoint was successfully reconciled
	ControlPlaneEndpointReadyCondition clusterv1.ConditionType = "ControlPlaneEnpointReady"
	// WaitingForControlPlaneEndpointReason (Severity=Warn) refers to a IntelCluster which is waiting for the control
	// plane endpoint to be populated through a ClusterConnection object
	WaitingForControlPlaneEndpointReason = "WaitingForControlPlaneEndpoint"
	// WaitingForControlPlaneEndpointReason (Severity=Error) refers to a IntelCluster which received a control plane
	// endpoint but with invalid values
	InvalidControlPlaneEndpointReason = "InvalidControlPlaneEndpoint"

	// WorkloadCreatedReadyCondition reports on whether a workload was successfully created with the infrastructure provider
	WorkloadCreatedReadyCondition clusterv1.ConditionType = "WorkloadCreatedReady"
	// WaitingForWorkloadToBeProvisonedReason (Severity=Info) refers to a IntelCluster which is waiting for
	// the workload to be created by the inventory provider
	WaitingForWorkloadToBeProvisonedReason = "WaitingForWorkloadToBeProvisoned"
	// InvalidWorkloadReason (Severity=Error) refers to a IntelCluster which received an invalid response from the
	// inventory provider when asked to create a workload
	InvalidWorkloadReason = "InvalidWorkload"
)
