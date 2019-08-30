package registry

const customDockerfileTmpl = `FROM {{ .BaseImage }}:{{ .Tag }}
{{range .DockerfileOptions -}}
{{ . }}
{{ end -}}
`
