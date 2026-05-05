// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	infrastructurev1alpha1 "github.com/open-edge-platform/cluster-api-provider-intel/api/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

var (
	// rke2CloudConfig is actual data["value"] taken from RKE2 bootstrap secret
	// here used for testing the parsing of cloud-config data in the controller
	rke2CloudConfig = []byte(base64.StdEncoding.EncodeToString([]byte(`#cloud-config
write_files:
  - path: /etc/rancher/rke2/config.yaml
    content: |
      tls-san:
      - example.local
`)))

	// k3sCloudConfig is actual data["value"] taken from K3S bootstrap secret
	// here used for testing the parsing of cloud-config data in the controller
	k3sCloudConfig = []byte(base64.StdEncoding.EncodeToString([]byte(`#cloud-config
write_files:
  - path: /etc/rancher/k3s/config.yaml
    owner: root:root
    permissions: "0640"
    content: |
      cluster-domain: cluster.edge
      cluster-init: true
runcmd:
  - echo success > /run/cluster-api/bootstrap-success.complete
`)))
)

func GetObjectRef(obj *metav1.ObjectMeta, kind string) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: infrastructurev1alpha1.GroupVersion.String(),
		Kind:       kind,
		Name:       obj.Name,
		Namespace:  obj.Namespace,
		UID:        obj.UID,
	}
}

func NewCluster(namespace, clusterName string) *clusterv1.Cluster {
	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: namespace,
		},
	}
	return cluster
}

func NewIntelCluster(namespace, intelClusterName, providerId string, cluster *clusterv1.Cluster) *infrastructurev1alpha1.IntelCluster { //nolint:lll
	return &infrastructurev1alpha1.IntelCluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      intelClusterName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: clusterv1.GroupVersion.String(),
					Kind:       "Cluster",
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
		Spec: infrastructurev1alpha1.IntelClusterSpec{
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "invalid host",
			},
			ProviderId: providerId,
		},
	}
}

func NewNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
func NewIntelClusterNoSpec(cluster *clusterv1.Cluster) *infrastructurev1alpha1.IntelCluster {
	return &infrastructurev1alpha1.IntelCluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: clusterv1.GroupVersion.String(),
					Kind:       "Cluster",
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
		Spec: infrastructurev1alpha1.IntelClusterSpec{
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "invalid host",
			},
		},
	}
}

func NewMachine(namespace, clusterName, machineName, bootstrapKind string) *clusterv1.Machine {
	bootstrapData := "bootstrap-data"
	machine := &clusterv1.Machine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Machine",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineName,
			Namespace: namespace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel:         clusterName,
				clusterv1.MachineControlPlaneLabel: "true",
			},
			UID: types.UID("1234567890"),
		},
		Spec: clusterv1.MachineSpec{
			ClusterName: clusterName,
			Bootstrap: clusterv1.Bootstrap{
				ConfigRef: clusterv1.ContractVersionedObjectReference{
					Kind: bootstrapKind,
					Name: "bootstrap-config",
				},
				DataSecretName: &bootstrapData,
			},
		},
	}
	return machine
}

func NewIntelMachine(namespace, intelMachineName string, machine *clusterv1.Machine) *infrastructurev1alpha1.IntelMachine { //nolint:lll
	return &infrastructurev1alpha1.IntelMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "IntelMachine",
			APIVersion: infrastructurev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      intelMachineName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: clusterv1.GroupVersion.String(),
					Kind:       "Machine",
					Name:       machine.Name,
					UID:        machine.UID,
				},
			},
			Labels: map[string]string{infrastructurev1alpha1.NodeGUIDKey: ""},
		},
	}
}

func NewIntelMachineBinding(namespace, intelMachineBindingName, nodeGUID, clusterName, machineTemplateName string) *infrastructurev1alpha1.IntelMachineBinding { //nolint:lll
	return &infrastructurev1alpha1.IntelMachineBinding{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      intelMachineBindingName,
			Namespace: namespace,
		},
		Spec: infrastructurev1alpha1.IntelMachineBindingSpec{
			NodeGUID:                 nodeGUID,
			ClusterName:              clusterName,
			IntelMachineTemplateName: machineTemplateName,
		},
	}
}

func NewIntelMachineTemplate(namespace, name, parent string) *infrastructurev1alpha1.IntelMachineTemplate { //nolint:lll
	annotations := map[string]string{}
	if parent != "" {
		annotations[clusterv1.TemplateClonedFromNameAnnotation] = parent
	}
	return &infrastructurev1alpha1.IntelMachineTemplate{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: infrastructurev1alpha1.IntelMachineTemplateSpec{},
	}
}

func NewRKE2BootstrapSecret(namespace, secretName string) *corev1.Secret {
	value, err := base64.StdEncoding.DecodeString(string(rke2CloudConfig))
	if err != nil {
		panic(err)
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"format": []byte("cloud-config"),
			"value":  value,
		},
	}
}

func NewK3SBootstrapSecret(namespace, secretName string) *corev1.Secret {
	value, err := base64.StdEncoding.DecodeString(string(k3sCloudConfig))
	if err != nil {
		panic(err)
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"value": value,
		},
	}
}
