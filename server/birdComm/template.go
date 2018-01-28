package main

const bgp = `{{range .}}
protocol bgp {{.Name}} from peers {
  description "{{.Description}}";
  neighbor {{.Address}} as {{.As}};
}
{{end}}
`

const static = `protocol static {
  {{range . -}}
  {{if .Nexthop -}}route {{.Prefix}}/{{.Mask}} via {{.Nexthop}};
  {{else -}}route {{.Prefix}}/{{.Mask}} unreachable;
  {{end}}
{{- end}}
}`
