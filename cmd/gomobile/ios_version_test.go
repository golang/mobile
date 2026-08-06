// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestParseIOSVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{input: "13", major: 13},
		{input: "13.0", major: 13},
		{input: "13.1", major: 13, minor: 1},
		{input: "13.1.1", major: 13, minor: 1, patch: 1},
		{input: "13.10", major: 13, minor: 10},
		{input: "14", major: 14},
		{input: "", wantErr: true},
		{input: "13.", wantErr: true},
		{input: ".1", wantErr: true},
		{input: "13..1", wantErr: true},
		{input: "13.1.1.1", wantErr: true},
		{input: "13.x", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			major, minor, patch, err := parseIOSVersion(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseIOSVersion(%q) = (%d, %d, %d, nil), want error", tc.input, major, minor, patch)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseIOSVersion(%q) returned unexpected error: %v", tc.input, err)
			}
			if major != tc.major || minor != tc.minor || patch != tc.patch {
				t.Errorf("parseIOSVersion(%q) = (%d, %d, %d), want (%d, %d, %d)", tc.input, major, minor, patch, tc.major, tc.minor, tc.patch)
			}
		})
	}
}

func TestValidateCatalystVersion(t *testing.T) {
	tests := []struct {
		platform   string
		iOSVersion string
		wantErr    bool
	}{
		{platform: "maccatalyst", iOSVersion: "13"},
		{platform: "maccatalyst", iOSVersion: "13.1.1"},
		{platform: "maccatalyst", iOSVersion: "14"},
		{platform: "maccatalyst", iOSVersion: "12.9.9", wantErr: true},
		{platform: "maccatalyst", iOSVersion: "13.x", wantErr: true},
		{platform: "ios", iOSVersion: "invalid"},
	}

	for _, tc := range tests {
		t.Run(tc.platform+"/"+tc.iOSVersion, func(t *testing.T) {
			err := validateCatalystVersion(tc.platform, tc.iOSVersion)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCatalystVersion(%q, %q) error = %v, wantErr %t", tc.platform, tc.iOSVersion, err, tc.wantErr)
			}
		})
	}
}
