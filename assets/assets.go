package assets

import "embed"

//go:embed templates/*.html
var TemplatesFS embed.FS

//go:embed static/dist/*
var StaticFS embed.FS

//go:embed migrations/sqlite/*.sql
var SqliteMigrationFS embed.FS
var SqliteMigrationDir string = "migrations/sqlite"

//go:embed migrations/postgres/*.sql
var PostgresMigrationFS embed.FS
var PostgresMigrationDir string = "migrations/postgres"

//go:embed migrations/mysql/*.sql
var MysqlMigrationFS embed.FS
var MysqlMigrationDir string = "migrations/mysql"
