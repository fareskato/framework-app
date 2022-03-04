package kabarda

import (
	"database/sql"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// OpenDB open a connection to db and return a connection pool
func (k *Kabarda) OpenDB(dbType, dsn string) (*sql.DB, error) {
	if dbType == "postgres" || dbType == "postgresql" {
		dbType = "pgx"
	}
	// connect using sql
	dbPool, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}
	// ping
	err = dbPool.Ping()
	if err != nil {
		return nil, err
	}
	return dbPool, nil
}
