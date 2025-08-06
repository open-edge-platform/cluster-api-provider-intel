// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/bnkamalesh/errors"
)

type SchemeType string

const (
	HttpScheme               SchemeType = "https://"
	defaultIstioQuitEndpoint string     = "http://localhost:15020/quitquitquit"
)

var (
	hostMatch = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
)

// IsUrl checks if the present
func IsUrl(u string) bool {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		log.Error().Msgf("It is not a valid url format) %v", err)
		return false
	}
	return true
}

// ConvertUrlToFilepath converts URL to File path
// example:
// before: https://registry.intel.com/hello
// after: registryintelcom
func ConvertUrlToSecretpath(urls string) (string, error) {
	log.Info().Msgf("ConvertUrlToFilepath: %v", urls)
	withScheme := strings.HasPrefix(urls, "http")
	var filepath string
	if withScheme {
		log.Info().Msgf("withScheme")
		up, err := url.Parse(urls)
		if err != nil {
			return "", errors.ValidationErr(err, "error parsing URL")
		}
		hostname := up.Host
		filepath = hostMatch.ReplaceAllString(hostname, "")
	} else {
		log.Info().Msgf("without scheme")
		if strings.Contains(urls, "/") {
			filepathArray := strings.Split(urls, "/")
			filepath = hostMatch.ReplaceAllString(filepathArray[0], "")
		} else {
			return "", errors.New("error parsing URL, URL doesn't contain /")
		}
	}

	return filepath, nil
}

// RetrieveFQDN inspect and remove the sheme of http protocol
// example:
// before: https://registry.intel.com/hello
// after: registry.intel.com
func RetrieveFQDN(urls string) (string, error) {
	log.Info().Msgf("ConvertUrlToFilepath: %v", urls)
	withScheme := strings.HasPrefix(urls, "http")

	var fqdn string

	if withScheme {
		log.Info().Msgf("withScheme")
		fqdn = urls
	} else {
		log.Info().Msgf("without scheme")
		if strings.Contains(urls, "/") {
			fqdnhArray := strings.Split(urls, "/")
			fqdn = "http://" + fqdnhArray[0]
		} else {
			fqdn = "http://" + urls
		}

	}

	up, err := url.Parse(fqdn)
	if err != nil {
		return "", errors.ValidationErr(err, "error parsing URL")
	}
	return up.Host + up.Path, nil
}

// Terminate Sidecar
// This is a WA to terminate Istio sidecar which block the kubernetes jobs complate
// https://github.com/istio/istio/issues/11659
// The issue was fixed in kubernetes 1.28

func TerminateSideCar(istioQuitEndpoint string) error {
	bodyReader := bytes.NewReader([]byte(``))

	if istioQuitEndpoint == "" {
		istioQuitEndpoint = defaultIstioQuitEndpoint
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, istioQuitEndpoint, bodyReader)
	if err != nil {
		log.Error().Err(err).Msg("failed to send quit to istio proxy.")
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to send quit to istio proxy.")
		return err
	}
	defer resp.Body.Close()
	log.Debug().Msgf("Close istio proxy with res %+v.", resp)
	return nil
}
