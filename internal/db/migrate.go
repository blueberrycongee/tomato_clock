package db

import "database/sql"

const schemaVersion = 1

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS settings (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL
    );`); err != nil {
		return err
	}

	var v int
	row := tx.QueryRow(`SELECT value FROM settings WHERE key = 'schema_version';`)
	if err := row.Scan(&v); err != nil {
		if err == sql.ErrNoRows {
			v = 0
		} else {
			return err
		}
	}

	if v < 1 {
		if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS task (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            note TEXT,
            is_done INTEGER NOT NULL DEFAULT 0,
            repeat_rule TEXT NOT NULL DEFAULT 'none',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`); err != nil {
			return err
		}

		if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS timer_session (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER,
            mode TEXT NOT NULL,
            target_seconds INTEGER NOT NULL DEFAULT 0,
            started_at DATETIME,
            ended_at DATETIME,
            interrupted INTEGER NOT NULL DEFAULT 0,
            FOREIGN KEY(task_id) REFERENCES task(id)
        );`); err != nil {
			return err
		}

		if _, err := tx.Exec(`INSERT OR REPLACE INTO settings(key, value) VALUES('schema_version', ?);`, schemaVersion); err != nil {
			return err
		}
	}

	return tx.Commit()
}
