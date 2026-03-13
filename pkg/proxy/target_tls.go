package proxy

import (
	"crypto/tls"
	"net/http"

	"github.com/rfancn/prism/pkg/types"
)

// createTransport creates an HTTP transport with TLS configuration.
func createTransport(targetTLS *types.TargetTLSConfig) *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // Default to secure
		},
	}

	if targetTLS != nil && targetTLS.InsecureSkipVerify {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	return transport
}