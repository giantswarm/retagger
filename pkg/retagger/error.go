package retagger

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var rateLimitedError = &microerror.Error{
	Kind: "rateLimitedError",
}

// IsRateLimited asserts rateLimitedError.
func IsRateLimited(err error) bool {
	return microerror.Cause(err) == rateLimitedError
}
