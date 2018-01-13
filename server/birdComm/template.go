package main

const config = `
{{ range . }}
protocol bgp {{ .Name }} {
  local as {{ .LocalAS }};
  neighbor {{ .Address }} as {{ .AS }};
  multihop;
  {{ if .Password }}password "{{ .Password }}";{{ end }}
  import filter bgp_in;
}

{{ end }}
`
