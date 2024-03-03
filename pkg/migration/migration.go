package migration

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Migrate(db *sql.DB, path string, isUp bool) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Println(err, "init kena error")
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		path,
		"postgres", driver)
	if err != nil {
		fmt.Println(err, "disini kena error")
		return err
	}
	if isUp {
		err = m.Up()
		if err != nil {
			return err
		}
	} else {
		err = m.Down()
		if err != nil {
			return err
		}
	}
	return nil
}
