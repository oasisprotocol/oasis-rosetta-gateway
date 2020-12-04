// Package common implements common things for Oasis Core Rosetta Gateway.
package common

import (
	"runtime"
	"strings"
)

var (
	// SoftwareVersion represents the Oasis Core Rosetta Gateway's version and
	// should be set by the linker.
	SoftwareVersion = "0.0.0-unset"

	// RosettaAPIVersion represents the Rosetta API version with which the
	// Oasis Core Rosetta Gateway is guaranteed to be compatible with.
	RosettaAPIVersion = "1.4.1"

	// ToolchainVersion is the version of the Go compiler/standard library.
	ToolchainVersion = strings.TrimPrefix(runtime.Version(), "go")
)
