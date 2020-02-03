package retagger

import (
	"context"
	"net/http"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/nokia/docker-registry-client/registry"
)

// backoffTransport retries requests on 429 responses
type backoffTransport struct {
	Transport http.RoundTripper
	logger    micrologger.Logger
}

// RoundTrip implements the RoundTripper interface for backoffTransport
// It acts as a sort of middleware, round-tripping using the wrapped transport,
// retrying on 429 error using exponential backoff and passing through on success
// or any other error or if the backoff retry time limit is reached.
func (t *backoffTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var innerErr error
	var resp *http.Response

	o := func() error {
		var respErr error
		resp, respErr = t.Transport.RoundTrip(request)
		// Internal error, return nil to prevent retry
		if respErr != nil {
			innerErr = respErr
			return nil
		}
		// Rate limited
		if resp.StatusCode == 429 {
			return rateLimitedError
		}
		// Not rate limited, return nil to prevent retry
		return nil
	}
	b := backoff.NewExponential(time.Minute, 10*time.Second)
	n := backoff.NewNotifier(t.logger, context.Background())
	backoffErr := backoff.RetryNotify(o, b, n)

	// Report errors unrelated to rate limiting first
	if innerErr != nil {
		return nil, microerror.Mask(innerErr)
	}
	// Rate limited and backoff time limit was reached
	if backoffErr != nil {
		return nil, microerror.Mask(backoffErr)
	}

	return resp, nil
}

// wrapTransport wraps the given transport with several custom RoundTrippers that handle
// - Token-based authentication (registry.TokenTransport)
// - HTTP basic authentication (registry.BasicTransport)
// - Rate limiting (backoffTransport)
// - Errors (registry.ErrorTransport)
func wrapTransport(transport http.RoundTripper, url string, logger micrologger.Logger) http.RoundTripper {
	transport = &registry.TokenTransport{
		Transport: transport,
	}
	transport = &registry.BasicTransport{
		Transport: transport,
		URL:       url,
	}
	transport = &backoffTransport{
		Transport: transport,
		logger:    logger,
	}
	transport = &registry.ErrorTransport{
		Transport: transport,
	}
	return transport
}
