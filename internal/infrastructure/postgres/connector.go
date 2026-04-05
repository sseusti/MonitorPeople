package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/lib/pq"
)

func OpenDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "3D000" {
			if createErr := ensureDatabaseExists(dsn); createErr != nil {
				return nil, createErr
			}

			db, err = sql.Open("postgres", dsn)
			if err != nil {
				return nil, err
			}
			if err := db.Ping(); err != nil {
				db.Close()
				return nil, err
			}
			return db, nil
		}
		return nil, err
	}

	return db, nil
}

func ensureDatabaseExists(dsn string) error {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return errors.New("database name is empty in DATABASE_URL")
	}

	adminURL := *parsedURL
	adminURL.Path = "/postgres"

	adminDB, err := sql.Open("postgres", adminURL.String())
	if err != nil {
		return err
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return err
	}

	_, err = adminDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "42P04" {
			return nil
		}
		return err
	}

	return nil
}
