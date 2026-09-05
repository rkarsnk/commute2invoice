// Package db は、SQLite データベースの初期化・接続管理を提供します。
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL embed.FS

// InitDB は、SQLite データベースを初期化し、接続を返します。
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrateSchema(db); err != nil {
		return nil, err
	}

	return db, nil
}

// migrateSchema は、schema.sql を読み込んで実行し、テーブルを作成します。
func migrateSchema(db *sql.DB) error {
	schemaBytes, err := fs.ReadFile(schemaSQL, "schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	schemaSQL := string(schemaBytes)

	_, err = db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close は、データベース接続を閉じます。
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
