package retagger

import (
	"os"
	"os/exec"
)

func Run(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
