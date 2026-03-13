// Package types provides shared types across packages.
package types

// TargetTLSConfig contains TLS configuration for target connections.
type TargetTLSConfig struct {
	InsecureSkipVerify bool
}