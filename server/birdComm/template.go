package main

const bgp = `{{range $k, $v := .}}
protocol bgp {{$v.Name}} from peers {
  {{with $v.Description}}description "{{.}}";{{end}}
  neighbor {{$k}} as {{$v.As}};
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
