package backoff

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

var testError = &microerror.Error{
	Kind: "testError",
}

// IsTestError asserts testError.
func IsTestError(err error) bool {
	return microerror.Cause(err) == testError
}

func Test_Permanent(t *testing.T) {
	var counter int
	var err error

	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	{
		o := func() error {
			if counter == 3 {
				return Permanent(testError)
			}
			counter++

			return fmt.Errorf("permanent usage failed error")
		}
		b := NewConstant(7*time.Second, 1*time.Second)
		n := NewNotifier(l, context.Background())

		err := RetryNotify(o, b, n)
		if IsTestError(err) {
			// fall through
		} else if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}
}
