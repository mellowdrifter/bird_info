package main

const bgp = `{{range .}}
protocol bgp {{.Name}} from peers {
  description "{{.Description}}";
  neighbor {{.Address}} as {{.As}};
}

{{end}}
`

const static = `protocol static {
  {{range .}}
  route {{.Route}}{{.Nexthop}}
  {{end}};
}`
