package registry

type Dockerfile struct {
	BaseImage         string
	Sha               string
	DockerfileOptions []string
	Tag               string
}
