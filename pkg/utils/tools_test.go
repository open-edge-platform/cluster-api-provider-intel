// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUrl(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success",
			args: args{
				u: "https://www.google.com",
			},
			want: true,
		},
		{
			name: "failure",
			args: args{
				u: "1234",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsUrl(tt.args.u), "IsUrl(%v)", tt.args.u)
		})
	}
}

func TestConvertUrlToSecretpath(t *testing.T) {
	type args struct {
		urls string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid URL with scheme",
			args: args{
				urls: "https://registry.intel.com/hello",
			},
			want:    "registryintelcom",
			wantErr: false,
		},
		{
			name: "Valid URL without scheme",
			args: args{
				urls: "registry.intel.com/hello",
			},
			want:    "registryintelcom",
			wantErr: false,
		},
		{
			name: "Invalid URL",
			args: args{
				urls: "12312312radfas@",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertUrlToSecretpath(tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertUrlToSecretpath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "ConvertUrlToSecretpath(%v)", tt.args.urls)
		})
	}
}

func TestRetrieveFQDN(t *testing.T) {
	type args struct {
		urls string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success with scheme",
			args: args{
				urls: "https://registry.intel.com",
			},
			want:    "registry.intel.com",
			wantErr: false,
		},
		{
			name: "success without scheme",
			args: args{
				urls: "registry.intel.com",
			},
			want:    "registry.intel.com",
			wantErr: false,
		},
		{
			name: "success without scheme",
			args: args{
				urls: "registry.intel.com/hello",
			},
			want:    "registry.intel.com",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RetrieveFQDN(tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveFQDN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "RetrieveFQDN(%v)", tt.args.urls)
		})
	}
}

func TestTerminateSideCarEmptyIstioEndpoint(t *testing.T) {

	err := TerminateSideCar("")
	assert.Error(t, err)

}

func TestTerminateSideCar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("https://registry.com/hello"))
	}))

	defer server.Close()

	err := TerminateSideCar(server.URL)
	assert.Equalf(t, err, nil, "No errors")

}
