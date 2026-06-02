package migrations

import "embed"

//go:embed sql/*.sql
var SQLfs embed.FS
