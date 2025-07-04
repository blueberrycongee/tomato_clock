package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var conn *sql.DB

func Conn() *sql.DB {
	return conn
}

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbPath := filepath.Join(home, ".tomato_clock.db")

	c, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	if err := c.Ping(); err != nil {
		return err
	}
	conn = c

	if err := migrate(c); err != nil {
		return err
	}
	return nil
}
