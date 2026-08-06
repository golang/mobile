// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func parseIOSVersion(s string) (major, minor, patch int, err error) {
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return 0, 0, 0, errors.New("must have one to three numeric components")
	}

	var components [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, 0, err
		}
		components[i] = n
	}

	return components[0], components[1], components[2], nil
}

func validateCatalystVersion(platform, iOSVersion string) error {
	if platform != "maccatalyst" {
		return nil
	}

	major, _, _, err := parseIOSVersion(iOSVersion)
	if err != nil {
		return fmt.Errorf("invalid iOS version %q: %v", iOSVersion, err)
	}
	if major < 13 {
		return errors.New("catalyst requires iOS version 13 or higher")
	}
	return nil
}
