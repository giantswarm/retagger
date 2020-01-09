package microerror

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
)

func Test_Stack(t *testing.T) {
	testCases := []struct {
		name                string
		inputErr            error
		expectedStackRegexp string
	}{
		{
			name:                "case 0: annotated microerror error",
			inputErr:            Maskf(&Error{Kind: "testKind"}, "annotation"),
			expectedStackRegexp: `^\[\{[a-zA-Z_/-]*/src/github.com/giantswarm/microerror/funcs_test.go:\d+: annotation\} \{test kind\}\]$`,
		},
		{
			name:                "case 1: non annotated microerror error",
			inputErr:            &Error{Kind: "testKind"},
			expectedStackRegexp: "^test kind$",
		},
		{
			name:                "case 2: external error",
			inputErr:            errors.New("external error"),
			expectedStackRegexp: "^external error$",
		},
		{
			name:                "case 3: nil",
			inputErr:            nil,
			expectedStackRegexp: "^<nil>$",
		},
	}

	for i, tc := range testCases {
		re, err := regexp.Compile(tc.expectedStackRegexp)
		if err != nil {
			t.Fatalf("err = %q, want nil", Stack(err))
		}

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stack := Stack(tc.inputErr)
			if !re.MatchString(stack) {
				t.Fatalf("stack = %q, want matching regexp %#q", stack, tc.expectedStackRegexp)
			}
		})
	}
}

func Test_Desc(t *testing.T) {
	testCases := []struct {
		name         string
		inputErr     error
		expectedDesc string
	}{
		{
			name:         "case 0: microerror without Desc",
			inputErr:     Maskf(&Error{Kind: "testKind"}, "annotation"),
			expectedDesc: "",
		},
		{
			name:         "case 1: microerror with Desc",
			inputErr:     &Error{Kind: "testKind", Desc: "test description"},
			expectedDesc: "test description",
		},
		{
			name:         "case 2: external error",
			inputErr:     errors.New("external error"),
			expectedDesc: "",
		},
		{
			name:         "case 3: nil",
			inputErr:     nil,
			expectedDesc: "",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			desc := Desc(tc.inputErr)
			if desc != tc.expectedDesc {
				t.Fatalf("desc = %q, expected %q", desc, tc.expectedDesc)
			}
		})
	}
}
