package main

import (
	"testing"

	"github.com/containers/image/v5/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var _ yaml.Unmarshaler = (*tlsVerifyConfig)(nil)

func TestTLSVerifyConfig(t *testing.T) {
	type container struct { // An example of a larger config file
		TLSVerify tlsVerifyConfig `yaml:"tls-verify"`
	}

	for _, c := range []struct {
		input    string
		expected tlsVerifyConfig
	}{
		{
			input:    `tls-verify: true`,
			expected: tlsVerifyConfig{skip: types.OptionalBoolFalse},
		},
		{
			input:    `tls-verify: false`,
			expected: tlsVerifyConfig{skip: types.OptionalBoolTrue},
		},
		{
			input:    ``, // No value
			expected: tlsVerifyConfig{skip: types.OptionalBoolUndefined},
		},
	} {
		config := container{}
		err := yaml.Unmarshal([]byte(c.input), &config)
		require.NoError(t, err, c.input)
		assert.Equal(t, c.expected, config.TLSVerify, c.input)
	}

	// Invalid input
	config := container{}
	err := yaml.Unmarshal([]byte(`tls-verify: "not a valid bool"`), &config)
	assert.Error(t, err)
}
