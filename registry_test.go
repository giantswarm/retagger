package main

import "testing"

func Test_GetAuthURL(t *testing.T) {
	tcs := []struct {
		challenge     string
		expectedURL   string
		expectedError bool
		description   string
	}{
		{
			description: "base case",
			challenge:   `Bearer realm="https://dockerauth.aliyuncs.com/auth",service="registry.aliyuncs.com:cn-shanghai:26888",scope="repository:giantswarm/build-test:pull"`,
			expectedURL: "https://dockerauth.aliyuncs.com/auth?service=registry.aliyuncs.com:cn-shanghai:26888&scope=repository:giantswarm/build-test:pull",
		},
		{
			description: "missing scope is allowed",
			challenge:   `Bearer realm="https://dockerauth.aliyuncs.com/auth",service="registry.aliyuncs.com:cn-shanghai:26888"`,
			expectedURL: "https://dockerauth.aliyuncs.com/auth?service=registry.aliyuncs.com:cn-shanghai:26888&scope=",
		},
		{
			description:   "missing Bearer causes error",
			challenge:     `realm="https://dockerauth.aliyuncs.com/auth",service="registry.aliyuncs.com:cn-shanghai:26888"`,
			expectedError: true,
		},
		{
			description:   "missing service and scope causes error",
			challenge:     `Bearer realm="https://dockerauth.aliyuncs.com/auth"`,
			expectedError: true,
		},
		{
			description:   "missing key value separator in realm causes error",
			challenge:     `Bearer realm:"https://dockerauth.aliyuncs.com/auth",service="registry.aliyuncs.com:cn-shanghai:26888",scope="repository:giantswarm/build-test:pull"`,
			expectedError: true,
		},
		{
			description:   "missing key value separator in service causes error",
			challenge:     `Bearer realm="https://dockerauth.aliyuncs.com/auth",service:"registry.aliyuncs.com:cn-shanghai:26888",scope="repository:giantswarm/build-test:pull"`,
			expectedError: true,
		},
		{
			description:   "missing key value separator in scope causes error",
			challenge:     `Bearer realm="https://dockerauth.aliyuncs.com/auth",service="registry.aliyuncs.com:cn-shanghai:26888",scope:"repository:giantswarm/build-test:pull"`,
			expectedError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := getAuthURL(tc.challenge)
			if tc.expectedError && err == nil {
				t.Fatalf("expected error didn't happen")
			}
			if !tc.expectedError && err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if actual != tc.expectedURL {
				t.Fatalf("unexpected URL, want %q, got %q", tc.expectedURL, actual)
			}
		})
	}
}
