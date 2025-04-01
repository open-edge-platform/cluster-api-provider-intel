// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package southboundconfig

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseInputArg tests arguments are parsed correctly
func TestParseInputArg(t *testing.T) {
	// TODO: refactor to use table drive test and avoid using default values (because they will be set without specifying the flags on the command line)
	os.Args = append(os.Args, fmt.Sprintf("--grpcAddr=%v", defaultGrpcAddr))
	os.Args = append(os.Args, fmt.Sprintf("--grpcPort=%v", defaultGrpcPort))
	os.Args = append(os.Args, fmt.Sprintf("--traceURL=%v", defaultTraceURL))

	cfg := ParseInputArg()

	assert.Equal(t, defaultGrpcAddr, cfg.GrpcAddr)
	assert.Equal(t, defaultGrpcPort, cfg.GrpcPort)
	assert.Equal(t, defaultTraceURL, cfg.TraceURL)
}

func TestValidateConfig(t *testing.T) {
	args := []struct {
		name    string
		input   *Config
		wantErr bool
	}{
		{
			name: "Test ValidateConfig success",
			input: &Config{
				GrpcAddr:             defaultGrpcAddr,
				GrpcPort:             defaultGrpcPort,
				ReadinessProbeGrpcEP: defaultReadinessProbeGrpcEP,
				EnableTracing:        true,
				TraceURL:             defaultTraceURL,
			},
			wantErr: false,
		},
		{
			name: "Test ValidateConfig Failure - Invalid port data type",
			input: &Config{
				GrpcAddr:             defaultGrpcAddr,
				GrpcPort:             "abc",
				ReadinessProbeGrpcEP: defaultReadinessProbeGrpcEP,
			},
			wantErr: true,
		},
		{
			name: "Test ValidateConfig Failure - Invalid port range",
			input: &Config{
				GrpcAddr:             defaultGrpcAddr,
				GrpcPort:             "-1",
				ReadinessProbeGrpcEP: defaultReadinessProbeGrpcEP,
			},
			wantErr: true,
		},
		{
			name: "Test ValidateConfig Failure - Invalid host",
			input: &Config{
				GrpcAddr:             "",
				ReadinessProbeGrpcEP: defaultReadinessProbeGrpcEP,
				GrpcPort:             defaultGrpcPort,
			},
			wantErr: true,
		},
		{
			name: "Test ValidateConfig Failure - Invalid Readiness Probe GprcEP",
			input: &Config{
				GrpcAddr:             defaultGrpcAddr,
				GrpcPort:             defaultGrpcPort,
				ReadinessProbeGrpcEP: "1.2.3",
			},
			wantErr: true,
		},
		{
			name:    "Test ValidateConfig Failure - nil config",
			input:   nil,
			wantErr: true,
		},
	}

	for _, v := range args {
		t.Run(v.name, func(t *testing.T) {
			err := ValidateConfig(v.input)
			if (err != nil) != v.wantErr {
				t.Errorf("ValidateConfig() err %v, wantErr %v", err, v.wantErr)
				return
			}
		})
	}
}
