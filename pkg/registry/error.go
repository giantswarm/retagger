package registry

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidStatusCodeError = &microerror.Error{
	Kind: "invalidStatusCode",
}

// IsInvalidStatusCode asserts invalidStatusCodeError.
func IsInvalidStatusCode(err error) bool {
	return microerror.Cause(err) == invalidStatusCodeError
}

var invalidTemplateError = &microerror.Error{
	Kind: "invalidTemplateError",
}

// IsInvalidTemplate asserts invalidTemplateError.
func IsInvalidTemplate(err error) bool {
	return microerror.Cause(err) == invalidTemplateError
}

var invalidArgumentError = &microerror.Error{
	Kind: "invalidArgumentError",
}

// IsInvalidArgument asserts invalidArgumentError.
func IsInvalidArgument(err error) bool {
	return microerror.Cause(err) == invalidArgumentError
}

var dockerError = &microerror.Error{
	Kind: "dockerError",
}

// IsDocker asserts dockerError.
func IsDocker(err error) bool {
	return microerror.Cause(err) == dockerError
}
