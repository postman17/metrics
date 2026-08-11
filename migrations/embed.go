// Package migrations встраивает SQL-файлы миграций через go:embed.
package migrations

import "embed"

// FS — встроенная файловая система с SQL-миграциями (go:embed *.sql).
//go:embed *.sql
var FS embed.FS
