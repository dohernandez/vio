package swagger

import _ "embed"

//go:embed service.swagger.json
// SwgJSON contains the service.swagger.json definition.
var SwgJSON []byte
