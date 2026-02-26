Copyright © 2024-2026 Apple Inc. and the Pkl project authors

This product includes libraries that may be distributed under a different license.

{{ range $index, $value := . }}
---
{{ .Name }}
{{ .LicenseName }} - {{ .LicenseURL }}

***

{{ .LicenseText }}
{{ end }}
