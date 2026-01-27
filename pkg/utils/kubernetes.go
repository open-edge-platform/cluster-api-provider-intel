// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/naughtygopher/errors"
	"github.com/open-edge-platform/cluster-api-provider-intel/pkg/logging"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	HELMNAMESPACEENV = "POD_NAMESPACE"
)

var log = logging.GetLogger("utils")

func NewInClusterClient() (kubernetes.Interface, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.InternalErr(err, "failed to acquire cluster config")
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.InternalErr(err, "failed to gen new k8s client")
	}
	return client, nil
}

type KubernetesAPI struct {
	Client kubernetes.Interface
}

func (kapi *KubernetesAPI) NewInClusterClient() error {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return errors.InternalErr(err, "failed to generate cluster config")
	}
	kapi.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.InternalErr(err, "failed to gen new k8s client")
	}
	return nil
}

func NewInClusterDynamicClient() (dynamic.Interface, error) {
	restconf, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.InternalErr(err, "failed to get restconf")
	}
	restConfigValue := *restconf

	// QPS indicates the maximum QPS to the master from this client.
	// Burst is maximum burst for throttle.
	// https://pkg.go.dev/k8s.io/client-go@v0.29.0/rest#Config
	restConfigValue.QPS = float32(10)
	restConfigValue.Burst = int(100)
	restconf = &restConfigValue

	dynamicClient, err := dynamic.NewForConfig(restconf)
	if err != nil {
		return nil, errors.InternalErr(err, "failed to get kube config")
	}
	return dynamicClient, nil
}
