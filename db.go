package rdsnap

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type db struct {
	*sql.DB
}

func connectDB(engine, user, password, host, database string, port int64) (*db, error) {
	var dsn string

	switch engine {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			user,
			password,
			host,
			port,
			database,
		)
	case "postgres":
		dsn = fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s",
			user,
			password,
			host,
			port,
			database,
		)
	}

	d, err := sql.Open(engine, dsn)
	if err != nil {
		return nil, err
	}

	return &db{d}, nil
}

func (d *db) ping() error {
	return d.Ping()
}

func (d *db) truncateTable(table string) error {
	_, err := d.Exec("TRUNCATE TABLE " + table)
	if err != nil {
		return err
	}

	return nil
}

func (d *db) close() {
	d.Close()
}
