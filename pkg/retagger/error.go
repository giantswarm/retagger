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

// Copied from net/http/transport.go:911
var trailerEOFErrorString = "http: unexpected EOF reading trailer"

// IsTrailerEOF asserts error string matches trailerEOFErrorString.
func IsTrailerEOF(err error) bool {
	return microerror.Cause(err).Error() == trailerEOFErrorString
}
