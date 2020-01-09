package microerror

import (
	"fmt"
	"regexp"
	"testing"
)

func Test_Error(t *testing.T) {
	var err error

	testError := &Error{
		Kind: "testError",
	}

	err = Mask(testError)

	got := fmt.Sprintf("%#v\n", err)
	r, err := regexp.Compile(`[.* test error]`)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
	if !r.MatchString(got) {
		t.Fatalf("expected %#v got %#v", true, false)
	}
}
