// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package mock_secret_client

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
)

type MockSecretClient struct{}

func (MockSecretClient) Initialize() error {
	return nil
}

func (MockSecretClient) Put(secretPath string, secretkvMap map[string]interface{}) error {
	log.Info().Msgf("PUT request received for %v", secretPath)
	return nil
}

func (MockSecretClient) Patch(secretPath string, secretkvMap map[string]interface{}) error {
	log.Info().Msgf("patch request received for %v", secretPath)
	return nil
}

func (MockSecretClient) Get(secretPath string) (*api.KVSecret, error) {
	switch secretPath {
	case "co-cm-db-pwd":
		return &api.KVSecret{
			Data: map[string]interface{}{
				"db-pwd": "password",
			},
			VersionMetadata: nil,
			CustomMetadata:  nil,
			Raw:             nil,
		}, nil
	case "test-repo":
		return &api.KVSecret{
			Data: map[string]interface{}{
				"cacert": "test-repo-ca",
			},
			VersionMetadata: nil,
			CustomMetadata:  nil,
			Raw:             nil,
		}, nil
	default:
		return nil, fmt.Errorf("unknown key %v", secretPath)
	}
}

func (MockSecretClient) Delete(secretPath string) error {
	log.Info().Msgf("delete request received for %v", secretPath)
	return nil
}

func (MockSecretClient) Undelete(secretPath string, versions []int) error {
	log.Info().Msgf("Undelete request received for %v", secretPath)
	return nil
}
