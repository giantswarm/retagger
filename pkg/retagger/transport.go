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

type backoffTransport struct {
	Transport http.RoundTripper
	logger    micrologger.Logger
}

func (t *backoffTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response

	{
		o := func() error {
			var respErr error
			resp, respErr = t.Transport.RoundTrip(request)
			// Internal error, return nil to prevent retry
			if respErr != nil {
				err = respErr
				return nil
			}
			// Rate limited
			if resp.StatusCode == 429 {
				return microerror.New("rate limited")
			}
			// Not rate limited, return nil to prevent retry
			return nil
		}
		b := backoff.NewExponential(time.Minute, 10*time.Second)
		n := backoff.NewNotifier(t.logger, context.Background())
		backoffErr := backoff.RetryNotify(o, b, n)
		// Report errors unrelated to rate limiting first
		if err != nil {
			return nil, microerror.Mask(err)
		}
		// Rate limited and backoff wasn't sufficient
		if backoffErr != nil {
			return nil, microerror.Mask(backoffErr)
		}
	}

	return resp, nil
}

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

