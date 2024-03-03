package psql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PSQL struct {
	Name              string `mapstructure:"name"`
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	UserName          string `mapstructure:"user_name"`
	Password          string `mapstructure:"password"`
	SSLMode           string `mapstructure:"ssl_mode"`
	MaxOpenConnection int    `mapstructure:"max_open_connection"`
	MaxIdleConnection int    `mapstructure:"max_idle_connection"`
	MaxLifeTime       int    `mapstructure:"max_lifetime"`
	DebugMode         bool   `mapstructure:"debug_mode"`
	MigrationPath     string `mapstructure:"migration_path"`
}

func PsqlConnect(conf PSQL) *sql.DB {
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.UserName, conf.Password, conf.Name, conf.SSLMode)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(conf.MaxOpenConnection)
	db.SetMaxIdleConns(conf.MaxIdleConnection)
	db.SetConnMaxLifetime(time.Duration(conf.MaxLifeTime) * time.Hour)
	boil.SetDB(db)
	boil.DebugMode = conf.DebugMode
	return db
}
