Copyright © 2024 Apple Inc. and the Pkl project authors

The pkl-gen-go binary includes libraries that may be distributed under a different license.

{{ range $index, $value := . }}
---
{{ .Name }}
{{ .LicenseName }} - {{ .LicenseURL }}

***

{{ .LicenseText }}
{{ end }}
