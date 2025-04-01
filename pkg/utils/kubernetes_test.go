// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	errInclusterConfig     = errors.New("generate in cluster config error")
	errInclusterClient     = errors.New("generate in cluster client error")
	kubeconfigGoodFilepath = "testdata/testkubeconfigcontent"
)

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func readkubeconfigTobyte(t *testing.T, filepath string) []byte {

	if b, err := os.ReadFile(filepath); err != nil {
		t.Error("[Fatal]: Cannot read file from filepath`" + filepath + "`.")
		return nil
	} else {
		return b
	}

}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func patchInClusterConfig(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(restclient.InClusterConfig, func() (*restclient.Config, error) {
		if fail == true {
			return nil, fmt.Errorf("in cluster config fail")
		}
		return &restclient.Config{}, nil
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
	return patch
}

func patchNewConfigConfig(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(kubernetes.NewForConfig, func(c *restclient.Config) (*kubernetes.Clientset, error) {
		if fail == true {
			return nil, fmt.Errorf("NewForConfig fail")
		}
		return &kubernetes.Clientset{}, nil
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
	return patch
}

func TestNewInClusterClient(t *testing.T) {
	cases := []struct {
		name           string
		wantErr        bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:           "NewInClusterClient InClusterConfig fail",
			wantErr:        true,
			funcBeforeTest: nil,
		},
		{
			name:    "NewInClusterClient NewForConfig fail",
			wantErr: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchInClusterConfig(t, false)
				patch2 := patchNewConfigConfig(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:    "NewInClusterClient success",
			wantErr: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchInClusterConfig(t, false)
				patch2 := patchNewConfigConfig(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			client, err := NewInClusterClient()

			if (err != nil) != tc.wantErr {
				t.Errorf("got err %v wantErr %v", err, tc.wantErr)
				return
			}
			if err == nil && client == nil {
				t.Errorf("client is nil")
				return
			}
		})
	}
}

func TestKubernetesAPI_NewInClusterClient(t *testing.T) {
	cases := []struct {
		name           string
		wantErr        bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:           "NewInClusterClient InClusterConfig fail",
			wantErr:        true,
			funcBeforeTest: nil,
		},
		{
			name:    "NewInClusterClient NewForConfig fail",
			wantErr: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchInClusterConfig(t, false)
				patch2 := patchNewConfigConfig(t, true)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
		{
			name:    "NewInClusterClient success",
			wantErr: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchInClusterConfig(t, false)
				patch2 := patchNewConfigConfig(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {
			var kapi KubernetesAPI
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := kapi.NewInClusterClient()

			if (err != nil) != tc.wantErr {
				t.Errorf("got err %v wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}

func TestNewInClusterDynamicClient(t *testing.T) {
	cases := []struct {
		name           string
		expectError    error
		funcBeforeTest func()
	}{
		{
			name:        "NewInClusterDynamicClient success",
			expectError: nil,
			funcBeforeTest: func() {
				patchInClusterConfigSuccess(t)
			},
		},
		{
			name:        "Generate incluster config fail",
			expectError: errInclusterConfig,
			funcBeforeTest: func() {
				patchInClusterConfigFail(t)
			},
		},
		{
			name:        "Generate incluster dynamic client fail",
			expectError: errInclusterClient,
			funcBeforeTest: func() {
				patchInClusterConfigSuccess(t)
				patchInClusterDynamicClientFail(t)
			},
		},
	}

	for _, tc := range cases {

		if tc.funcBeforeTest != nil {
			tc.funcBeforeTest()
		}

		_, err := NewInClusterDynamicClient()

		if !isExpectedError(err, tc.expectError) {
			t.Errorf("Unexpected error: %v", err)
		}

	}
}

func patchInClusterDynamicClientFail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(dynamic.NewForConfig, func(inConfig *restclient.Config) (*dynamic.DynamicClient, error) {
		unpatch(t, patch)
		return nil, errInclusterClient
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchInClusterConfigSuccess(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(restclient.InClusterConfig, func() (*restclient.Config, error) {
		unpatch(t, patch)
		context := readkubeconfigTobyte(t, kubeconfigGoodFilepath)
		clientCfg, err := clientcmd.NewClientConfigFromBytes(context)
		if err != nil {
			return nil, err
		}
		restCfg, err := clientCfg.ClientConfig()
		if err != nil {
			return nil, err
		}
		return restCfg, nil
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}

func patchInClusterConfigFail(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(restclient.InClusterConfig, func() (*restclient.Config, error) {
		unpatch(t, patch)
		return nil, errInclusterConfig
	})
	if patchErr != nil {
		t.Errorf("patch error %v", patchErr)
	}
}
