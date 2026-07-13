package sqlite

type ClientOptions struct {
	Path            string
	ForeignKeys     bool
	BusyTimeoutMS   int
	MaxOpenConns    int
	MaxIdleConns    int
	SchemaPath      string
	RunMigrations   bool
	MigrationsTable string
	ConnectionName  string
}

type MigrationRecord struct {
	ID        int64
	Name      string
	Checksum  string
	AppliedAt string
}
