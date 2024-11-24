package swagger

import (
	_ "embed"
)

// SwgJSON contains the service.swagger.json definition.
//
//go:embed service.swagger.json
var SwgJSON []byte
