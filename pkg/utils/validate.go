// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"net"
	url2 "net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	HttpSchemePath = "https://"
	// accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	hostnameRegexStringRFC1123 = `^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*?$`
)

// IsValidIPV4Port assumes the uRL is of the format <ip>:<port> for a valid url.
// The other assumption is that the host is a valid IP4
// TODO: Enhance this function later if other formats are valid
func IsValidIPV4Port(uRL string) error {
	out := strings.Split(uRL, ":")
	if len(out) != 2 {
		return fmt.Errorf("invalid url format: %v", uRL)
	}
	// validate ip
	if net.ParseIP(out[0]) == nil {
		return fmt.Errorf("invalid ip address in the url: %v", uRL)
	}
	// validate port
	if err := IsValidPort(out[1]); err != nil {
		return err
	}
	return nil
}

// IsValidHost TODO: This is a stupid validation. Fix this ASAP
func IsValidHost(host string) error {
	if host == "" {
		return fmt.Errorf("host is nil")
	}
	return nil
}

func IsValidPort(port string) error {
	p, err := strconv.Atoi(port)
	// validate port format
	if err != nil {
		return fmt.Errorf("invalid port format: %v", port)
	}
	// validate port range
	if err = IsValidPortInt(p); err != nil {
		return err
	}

	return nil
}

func IsValidPortInt(port int) error {
	// validate port range
	if port > 65535 || port < 1 {
		return fmt.Errorf("invalid port range: %v", port)
	}

	return nil
}

func IsValidDNSName(dnsName string) error {
	var domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
	if !domainRegexp.MatchString(dnsName) {
		return fmt.Errorf("%v is not a valid dns name", dnsName)
	}
	return nil
}

// IsValidHostnamePort validates a <dns>:<port> combination for fields typically used for socket address.
func IsValidHostnamePort(val string) bool {
	host, port, err := net.SplitHostPort(val)
	if err != nil {
		return false
	}
	var portNum int64
	// Port must be any port <= 65535.
	if portNum, err = strconv.ParseInt(port, 10, 32); err != nil || portNum > 65535 || portNum < 1 {
		return false
	}

	// If host is specified, it should match a DNS name
	if host != "" {
		if !regexp.MustCompile(hostnameRegexStringRFC1123).MatchString(host) {
			return false
		}
	}
	return true
}

func IsValidUrl(url string) error {
	withScheme := strings.HasPrefix(url, "http")
	var cto string
	if withScheme {
		cto = url
	} else {
		cto = HttpSchemePath + url
	}
	if _, err := url2.ParseRequestURI(cto); err != nil {
		return err
	}
	return nil
}

// IsValidNamespace  TODO: fix this rudimentary validation
func IsValidNamespace(ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace not specified")
	}
	return nil
}

func IsAbsFilePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("%s is not absolute file path", path)
	}
	return nil
}
