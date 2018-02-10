package main

const bgp = `{{range $k, $v := .Peer}}
protocol bgp {{$v.Name}} from peers {
  description "{{$v.Description}}";
  neighbor {{$v.Address}} as {{$v.As}};
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
