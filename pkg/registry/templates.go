package registry

const customDockerfileTmpl = `FROM {{ .BaseImage }}@sha256:{{ .Sha }}
{{range .DockerfileOptions -}}
{{ . }}
{{ end -}}
`
