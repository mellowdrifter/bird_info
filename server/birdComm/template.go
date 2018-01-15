package main

const bgp = `router id 1.1.1.1;
protocol device {}

protocol kernel {
  metric 64;
  import none;
}

template bgp peers {
  local as 100;
  multihop;
  password "password123";
}

{{range .}}
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
