package registry

import "fmt"

func TempDockerfileName(destinationTag string) string {
	return fmt.Sprintf("Dockerfile-%s", destinationTag)
}
