package migrations

import "embed"

// FS contains the SQL migrations used during service startup.
//
//go:embed *.sql
var FS embed.FS
