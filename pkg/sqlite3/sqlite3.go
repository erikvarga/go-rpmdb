package sqlite3

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"os"
	"slices"

	dbi "github.com/erikvarga/go-rpmdb/pkg/db"
	"golang.org/x/xerrors"
)

type SQLite3 struct {
	*sql.DB
}

var (
	// https://www.sqlite.org/fileformat.html
	SQLite3_HeaderMagic = []byte("SQLite format 3\x00")
	ErrorInvalidSQLite3 = xerrors.Errorf("invalid or unsupported SQLite3 format")
)

func Open(path string) (*SQLite3, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b := make([]byte, 16)
	if err = binary.Read(file, binary.LittleEndian, b); err != nil {
		return nil, xerrors.Errorf("binary read error: %w", err)
	}

	if !bytes.Equal(b, SQLite3_HeaderMagic) {
		return nil, ErrorInvalidSQLite3
	}

	// Prefer the "sqlite" driver from modernc.org/sqlite,
	// but fall back to "sqlite3" driver from github.com/mattn/go-sqlite3
	driver := "sqlite"
	if !slices.Contains(sql.Drivers(), driver) {
		driver = "sqlite3"
	}

	db, err := sql.Open(driver, path)
	if err != nil {
		return nil, xerrors.Errorf("failed to open sqlite3: %w", err)
	}

	return &SQLite3{db}, nil
}

func (db *SQLite3) Read(ctx context.Context) <-chan dbi.Entry {
	entries := make(chan dbi.Entry)

	go func() {
		defer close(entries)

		rows, err := db.QueryContext(ctx, "SELECT blob FROM Packages")
		if err != nil {
			_ = db.Close()
			entries <- dbi.Entry{
				Err: xerrors.Errorf("failed to SELECT query: %w", err),
			}
			return
		}

		if rows == nil {
			_ = db.Close()
			entries <- dbi.Entry{
				Err: xerrors.Errorf("query failed to return rows: %w", err),
			}
			return
		}

		defer func() {
			_ = rows.Close()
			if err := db.Close(); err != nil {
				entries <- dbi.Entry{
					Err: xerrors.Errorf("failed to close DB: %w", err),
				}
			}
		}()

		for rows.Next() {
			var blob string
			if err := rows.Scan(&blob); err != nil {
				entries <- dbi.Entry{
					Err: xerrors.Errorf("failed to Scan Row: %w", err),
				}
				return
			}

			select {
			case entries <- dbi.Entry{
				Value: []byte(blob),
				Err:   nil,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return entries
}
