// Package common implements common things for Oasis Core Rosetta Gateway.
package common

import (
	"runtime"
	"runtime/debug"
	"strings"

	rosettaTypes "github.com/coinbase/rosetta-sdk-go/types"
	ocVersion "github.com/oasisprotocol/oasis-core/go/common/version"
)

// VersionUnknown represents an unknown version.
const VersionUnknown = "unknown"

var (
	// SoftwareVersion represents the Oasis Core Rosetta Gateway's version and
	// should be set by the linker.
	SoftwareVersion = "0.0.0-unset"

	// RosettaAPIVersion represents the Rosetta API version with which the
	// Oasis Core Rosetta Gateway is guaranteed to be compatible with.
	RosettaAPIVersion = rosettaTypes.RosettaAPIVersion

	// ToolchainVersion is the version of the Go compiler/standard library.
	ToolchainVersion = strings.TrimPrefix(runtime.Version(), "go")
)

// GetOasisCoreVersion returns the version of the Oasis Core dependency or
// unknown if it can't be obtained.
func GetOasisCoreVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return VersionUnknown
	}

	for _, dep := range bi.Deps {
		if dep.Path == "github.com/oasisprotocol/oasis-core/go" {
			// Convert Go Modules compatible version to Oasis Core's canonical
			// version.
			return ocVersion.ConvertGoModulesVersion(dep.Version)
		}
	}
	return VersionUnknown
}
