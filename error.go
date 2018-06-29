package main

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidStatusCodeError = microerror.New("invalid status code")

// IsInvalidStatusCode asserts invalidStatusCodeError.
func IsInvalidStatusCode(err error) bool {
	return microerror.Cause(err) == invalidStatusCodeError
}

var invalidAuthenticateChallengeError = microerror.New("invalid authenticate challenge")

// IsInvalidAuthenticateChallenge asserts invalidAuthenticateChallengeError.
func IsInvalidAuthenticateChallenge(err error) bool {
	return microerror.Cause(err) == invalidAuthenticateChallengeError
}
