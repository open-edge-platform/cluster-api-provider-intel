// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
)

func TestIsHostnamePort(t *testing.T) {

	type testInput struct {
		data     string
		expected bool
	}
	testData := []testInput{
		{"bad..domain.name:234", false},
		{"extra.dot.com.", false},
		{"localhost:1234", true},
		{"192.168.1.1:1234", true},
		{":1234", true},
		{"domain.com:1334", true},
		{"this.domain.com:234", true},
		{"domain:75000", false},
		{"missing.port", false},
		{"https://google.com:1234", false},
		{"http://google.com:1234", false},
		{"http://google.com", false},
	}
	for _, td := range testData {
		res := IsValidHostnamePort(td.data)
		if td.expected != res {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.expected, res)
		}
	}
}

func TestIsValidIPPort(t *testing.T) {

	type testInput struct {
		data    string
		wantErr bool
	}
	testData := []testInput{
		{"192.168.1.1:1234", false},
		{":1234", true},
		{"domain:75000", true},
		{"1.2.3.4:75000", true},
		{"1.2.3.4:abc", true},
		{"missing.port", true},
		{"https://google.com:1234", true},
		{"http://google.com:1234", true},
		{"localhost:1234", true},
		{"0.0.0.0:1234", false},
		{"0.0.0.0", true},
		{"1.2.3.4", true},
		{"0.0.0.0:1234", false},
		{"::ffff:192.0.2.1", true},
		{"::ffff:192.0.2.1:1234", true},
	}
	for _, td := range testData {
		err := IsValidIPV4Port(td.data)
		if (err != nil) != td.wantErr {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.wantErr, err)
		}
	}
}

func TestIsValidHost(t *testing.T) {
	type testInput struct {
		data    string
		wantErr bool
	}
	testData := []testInput{
		{"192.168.1.1:1234", false},
		{":1234", false},
		{":1234", false},
		{"", true},
	}
	for _, td := range testData {
		err := IsValidHost(td.data)
		if (err != nil) != td.wantErr {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.wantErr, err)
		}
	}
}

func TestIsValidDNSName(t *testing.T) {
	type testInput struct {
		data    string
		wantErr bool
	}
	testData := []testInput{
		{"google.com", false},
		{"something", true},
		{"", true},
		{"1234", true},
		{"_123", true},
		{"localhost:1234", true},
	}
	for _, td := range testData {
		err := IsValidDNSName(td.data)
		if (err != nil) != td.wantErr {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.wantErr, err)
		}
	}
}

func TestIsValidUrl(t *testing.T) {
	type testInput struct {
		data    string
		wantErr bool
	}
	testData := []testInput{
		{"google.com", false},
		/* // TODO: How are these valid URLs!! The test isn't failing with these inputs
		{"1234", true},
		{"", true},
		*/
	}
	for _, td := range testData {
		err := IsValidUrl(td.data)
		if (err != nil) != td.wantErr {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.wantErr, err)
		}
	}
}

func TestIsValidNamespace(t *testing.T) {
	type testInput struct {
		data    string
		wantErr bool
	}
	testData := []testInput{
		{"", true},
		{"something", false},
	}
	for _, td := range testData {
		err := IsValidNamespace(td.data)
		if (err != nil) != td.wantErr {
			t.Errorf("Test failed for data: %v, want: %v got: %v", td.data, td.wantErr, err)
		}
	}
}

func TestIsAbsFilePath(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/programs/course1/hello1.go",
			wantErr: false,
		},
		{
			name:    "invalid path - case 1",
			path:    "../programs/course1/hello1.go",
			wantErr: true,
		},
		{
			name:    "invalid path - case 2",
			path:    "C:/programs/course1/hello1.go",
			wantErr: true,
		},
	}

	for _, tc := range cases {

		err := IsAbsFilePath(tc.path)

		if (err != nil) != tc.wantErr {
			t.Errorf("want: %v, got: %v", tc.wantErr, err)
		}
	}
}
