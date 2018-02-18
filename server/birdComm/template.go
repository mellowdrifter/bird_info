package main

const bgp = `{{range $k, $v := . -}}
protocol bgp {{$v.Name}} from peers {
  neighbor {{$k}} as {{$v.As}};
  {{with $v.Description}}description "{{.}}";{{end}}
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
