package backoff

import (
	"fmt"
	"testing"
	"time"
)

func Test_MaxRetries(t *testing.T) {
	var c int
	o := func() error {
		c++
		return fmt.Errorf("test error")
	}
	b := NewMaxRetries(3, 1*time.Second)

	s := time.Now()
	Retry(o, b)

	since := time.Since(s)
	if since > 3*time.Second {
		t.Fatalf("expected less than %d seconds got %f", 3, since.Seconds())
	}
	if c != 3 {
		t.Fatalf("expected %d got %d", 3, c)
	}
}
