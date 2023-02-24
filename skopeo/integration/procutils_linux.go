package main

import (
	"os/exec"
	"syscall"
)

// cmdLifecyleToParentIfPossible is a thin wrapper around prctl(PR_SET_PDEATHSIG)
// on Linux.
func cmdLifecycleToParentIfPossible(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
}
