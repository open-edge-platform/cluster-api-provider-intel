// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	tests := []struct {
		wantLevel   zerolog.Level
		lowerLevel  zerolog.Level
		higherLevel zerolog.Level
	}{
		{
			wantLevel:   zerolog.FatalLevel,
			lowerLevel:  zerolog.ErrorLevel,
			higherLevel: zerolog.FatalLevel,
		},
		{
			wantLevel:   zerolog.PanicLevel,
			lowerLevel:  zerolog.ErrorLevel,
			higherLevel: zerolog.PanicLevel,
		},
		{
			wantLevel:   zerolog.ErrorLevel,
			lowerLevel:  zerolog.WarnLevel,
			higherLevel: zerolog.PanicLevel,
		},
		{
			wantLevel:   zerolog.WarnLevel,
			lowerLevel:  zerolog.InfoLevel,
			higherLevel: zerolog.ErrorLevel,
		},
		{
			wantLevel:   zerolog.InfoLevel,
			lowerLevel:  zerolog.DebugLevel,
			higherLevel: zerolog.WarnLevel,
		},
	}
	for _, tt := range tests {
		// Set the global log level
		if err := flag.Set("logLevel", tt.wantLevel.String()); err != nil {
			t.Errorf("error setting level")
			continue
		}
		// Get the logger
		l := GetLogger("test")
		// Verify the log level is what is set
		if l.GetLevel() != tt.wantLevel {
			t.Errorf("want level: %v, got level: %v", tt.wantLevel, l.GetLevel())
			continue
		}
		// Check that lower log level is not enabled. Fail the case if enabled...
		if l.WithLevel(tt.lowerLevel).Enabled() {
			t.Errorf("this log level should not have enabled, level: %v", tt.lowerLevel)
			continue
		}

		// Check that current level and higher levels are enabled.
		if !l.WithLevel(tt.wantLevel).Enabled() || !l.WithLevel(tt.higherLevel).Enabled() {
			t.Errorf("level should have been enabled but it is not, cLevel: %v, hLevel: %v clEnabled: %v, hlEnabled: %v",
				tt.wantLevel, tt.higherLevel, l.WithLevel(tt.wantLevel).Enabled(), l.WithLevel(tt.higherLevel).Enabled())
			continue
		}
	}
}

func TestGetLoggerWithSetenvHuman(t *testing.T) {
	tests := []struct {
		wantLevel   zerolog.Level
		lowerLevel  zerolog.Level
		higherLevel zerolog.Level
	}{
		{
			wantLevel:   zerolog.FatalLevel,
			lowerLevel:  zerolog.ErrorLevel,
			higherLevel: zerolog.FatalLevel,
		},
	}

	os.Setenv("HUMAN", "true")
	for _, tt := range tests {
		// Set the global log level
		if err := flag.Set("logLevel", tt.wantLevel.String()); err != nil {
			t.Errorf("error setting level")
			continue
		}
		// Get the logger
		l := GetLogger("test")
		// Verify the log level is what is set
		if l.GetLevel() != tt.wantLevel {
			t.Errorf("want level: %v, got level: %v", tt.wantLevel, l.GetLevel())
			continue
		}
		// Check that lower log level is not enabled. Fail the case if enabled...
		if l.WithLevel(tt.lowerLevel).Enabled() {
			t.Errorf("this log level should not have enabled, level: %v", tt.lowerLevel)
			continue
		}

		// Check that current level and higher levels are enabled.
		if !l.WithLevel(tt.wantLevel).Enabled() || !l.WithLevel(tt.higherLevel).Enabled() {
			t.Errorf("level should have been enabled but it is not, cLevel: %v, hLevel: %v clEnabled: %v, hlEnabled: %v",
				tt.wantLevel, tt.higherLevel, l.WithLevel(tt.wantLevel).Enabled(), l.WithLevel(tt.higherLevel).Enabled())
			continue
		}
	}
}

func TestCallerMarshal(t *testing.T) {
	expectedString := "http_server_options.go:100"
	out := callerMarshal(0, "/Users/$USER/Workspace/$PACKAGE_NAME/pkg/$PATH_TO_FILE/http_server_options.go", 100)

	assert.Equal(t, expectedString, out)
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCLogger_TraceCtx(t *testing.T) {
	l := MCLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.TraceCtx(context.Background())
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCLogger_McSec(t *testing.T) {
	l := MCLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McSec()
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCCtxLogger_McSec(t *testing.T) {
	l := MCCtxLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McSec()
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCLogger_McErr(t *testing.T) {
	l := MCLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McErr(fmt.Errorf("dummy"))
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCCtxLogger_McErr(t *testing.T) {
	l := MCCtxLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McErr(fmt.Errorf("dummy"))
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCLogger_McError(t *testing.T) {
	l := MCLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McError("dummy")
}

// Primarily written for improving UT Coverage metrics. Nothing to test and verify at this time.
func TestMCCtxLogger_McError(t *testing.T) {
	l := MCCtxLogger{
		Logger: zerolog.Logger{},
	}
	_ = l.McError("dummy")
}
