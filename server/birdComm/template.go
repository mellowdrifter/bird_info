package main

const bgp = `{{range .}}
protocol bgp {{.Name}} from {{.Group}} {
  description "{{.Description}}";
  neighbor {{.Address}} as {{.AS}};
}

{{end}}
`

const static = `protocol static {
  {{range .}}
  route {{.Route}}{{.Nexthop}}
  {{end}};
}`
