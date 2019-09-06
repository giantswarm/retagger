package registry

type Dockerfile struct {
	BaseImage         string
	DockerfileOptions []string
	Tag               string
}
