
- repo:
  - src-openeuler/{{.PkgName}}
  committers:
  {{- range .Committers }}
  - openeuler_id: {{.OpeneulerId}}
    name: {{.Name}}
    email: {{.Email}}
  {{- end}}
