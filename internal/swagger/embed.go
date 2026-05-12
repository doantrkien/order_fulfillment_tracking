package swagger

import _ "embed"

//go:embed swagger.yaml
var Spec []byte
